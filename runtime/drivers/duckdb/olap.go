package duckdb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Create instruments
var (
	meter                 = otel.Meter("github.com/rilldata/rill/runtime/drivers/duckdb")
	queriesCounter        = observability.Must(meter.Int64Counter("queries"))
	queueLatencyHistogram = observability.Must(meter.Int64Histogram("queue_latency", metric.WithUnit("ms")))
	queryLatencyHistogram = observability.Must(meter.Int64Histogram("query_latency", metric.WithUnit("ms")))
	totalLatencyHistogram = observability.Must(meter.Int64Histogram("total_latency", metric.WithUnit("ms")))
)

func (c *connection) Dialect() drivers.Dialect {
	return drivers.DialectDuckDB
}

func (c *connection) WithConnection(ctx context.Context, priority int, fn drivers.WithConnectionFunc) error {
	// Check not nested
	if connFromContext(ctx) != nil {
		panic("nested WithConnection")
	}

	// Acquire connection
	conn, release, err := c.acquireOLAPConn(ctx, priority)
	if err != nil {
		return err
	}
	defer func() { _ = release() }()

	// Call fn with connection embedded in context
	wrappedCtx := contextWithConn(ctx, conn)
	ensuredCtx := contextWithConn(context.Background(), conn)
	return fn(wrappedCtx, ensuredCtx, conn.Conn)
}

func (c *connection) Exec(ctx context.Context, stmt *drivers.Statement) error {
	res, err := c.Execute(ctx, stmt)
	if err != nil {
		return err
	}
	if stmt.DryRun {
		return nil
	}
	err = res.Close()
	return c.checkErr(err)
}

func (c *connection) Execute(ctx context.Context, stmt *drivers.Statement) (res *drivers.Result, outErr error) {
	// We use the meta conn for dry run queries
	if stmt.DryRun {
		conn, release, err := c.acquireMetaConn(ctx)
		if err != nil {
			return nil, err
		}
		defer func() { _ = release() }()

		// TODO: Find way to validate with args

		name := uuid.NewString()
		_, err = conn.ExecContext(ctx, fmt.Sprintf("CREATE TEMPORARY VIEW %q AS %s", name, stmt.Query))
		if err != nil {
			return nil, c.checkErr(err)
		}

		_, err = conn.ExecContext(context.Background(), fmt.Sprintf("DROP VIEW %q", name))
		return nil, c.checkErr(err)
	}

	// Gather metrics only for actual queries
	var acquiredTime time.Time
	acquired := false
	start := time.Now()
	defer func() {
		totalLatency := time.Since(start).Milliseconds()
		queueLatency := acquiredTime.Sub(start).Milliseconds()

		attrs := attribute.NewSet(
			attribute.String("db", c.config.DBFilePath),
			attribute.Bool("cancelled", errors.Is(outErr, context.Canceled)),
			attribute.Bool("failed", outErr != nil),
		)

		queriesCounter.Add(ctx, 1, metric.WithAttributeSet(attrs))
		queueLatencyHistogram.Record(ctx, queueLatency, metric.WithAttributeSet(attrs))
		totalLatencyHistogram.Record(ctx, totalLatency, metric.WithAttributeSet(attrs))
		if acquired {
			// Only track query latency when not cancelled in queue
			queryLatencyHistogram.Record(ctx, totalLatency-queueLatency, metric.WithAttributeSet(attrs))
		}
	}()

	// Acquire connection
	conn, release, err := c.acquireOLAPConn(ctx, stmt.Priority)
	acquiredTime = time.Now()
	if err != nil {
		return nil, err
	}
	acquired = true

	// NOTE: We can't just "defer release()" because release() will block until rows.Close() is called.
	// We must be careful to make sure release() is called on all code paths.

	var cancelFunc context.CancelFunc
	if stmt.ExecutionTimeout != 0 {
		ctx, cancelFunc = context.WithTimeout(ctx, stmt.ExecutionTimeout)
	}

	rows, err := conn.QueryxContext(ctx, stmt.Query, stmt.Args...)
	if err != nil {
		if cancelFunc != nil {
			cancelFunc()
		}

		// err must be checked before release
		err = c.checkErr(err)
		_ = release()
		return nil, err
	}

	schema, err := rowsToSchema(rows)
	if err != nil {
		if cancelFunc != nil {
			cancelFunc()
		}

		// err must be checked before release
		err = c.checkErr(err)
		_ = rows.Close()
		_ = release()
		return nil, err
	}

	res = &drivers.Result{Rows: rows, Schema: schema}
	res.SetCleanupFunc(func() error {
		if cancelFunc != nil {
			cancelFunc()
		}

		return release()
	})

	return res, nil
}

func (c *connection) EstimateSize() (int64, bool) {
	var paths []string
	path := c.config.DBFilePath
	if path == "" {
		return 0, true
	}

	// Add .wal file path (e.g final size will be sum of *.db and *.db.wal)
	dbWalPath := fmt.Sprintf("%s.wal", path)
	paths = append(paths, path, dbWalPath)
	return fileSize(paths), true
}

func rowsToSchema(r *sqlx.Rows) (*runtimev1.StructType, error) {
	if r == nil {
		return nil, nil
	}

	cts, err := r.ColumnTypes()
	if err != nil {
		return nil, err
	}

	fields := make([]*runtimev1.StructType_Field, len(cts))
	for i, ct := range cts {
		nullable, ok := ct.Nullable()
		if !ok {
			nullable = true
		}

		t, err := databaseTypeToPB(ct.DatabaseTypeName(), nullable)
		if err != nil {
			return nil, err
		}

		fields[i] = &runtimev1.StructType_Field{
			Name: ct.Name(),
			Type: t,
		}
	}

	return &runtimev1.StructType{Fields: fields}, nil
}

func fileSize(paths []string) int64 {
	var size int64
	for _, path := range paths {
		if info, err := os.Stat(path); err == nil { // ignoring error since only error possible is *PathError
			size += info.Size()
		}
	}
	return size
}
