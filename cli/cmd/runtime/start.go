package runtime

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/redis/go-redis/v9"
	"github.com/rilldata/rill/cli/pkg/config"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"

	// Load connectors and reconcilers for runtime
	_ "github.com/rilldata/rill/runtime/drivers/bigquery"
	_ "github.com/rilldata/rill/runtime/drivers/druid"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	_ "github.com/rilldata/rill/runtime/drivers/gcs"
	_ "github.com/rilldata/rill/runtime/drivers/github"
	_ "github.com/rilldata/rill/runtime/drivers/https"
	_ "github.com/rilldata/rill/runtime/drivers/postgres"
	_ "github.com/rilldata/rill/runtime/drivers/s3"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	_ "github.com/rilldata/rill/runtime/reconcilers"
)

// Config describes runtime server config derived from environment variables.
// Env var keys must be prefixed with RILL_RUNTIME_ and are converted from snake_case to CamelCase.
// For example RILL_RUNTIME_HTTP_PORT is mapped to Config.HTTPPort.
type Config struct {
	HTTPPort            int                    `default:"8080" split_words:"true"`
	GRPCPort            int                    `default:"9090" split_words:"true"`
	LogLevel            zapcore.Level          `default:"info" split_words:"true"`
	MetricsExporter     observability.Exporter `default:"prometheus" split_words:"true"`
	TracesExporter      observability.Exporter `default:"" split_words:"true"`
	MetastoreDriver     string                 `default:"sqlite" split_words:"true"`
	MetastoreURL        string                 `default:"file:rill?mode=memory&cache=shared" split_words:"true"`
	AllowedOrigins      []string               `default:"*" split_words:"true"`
	AuthEnable          bool                   `default:"false" split_words:"true"`
	AuthIssuerURL       string                 `default:"" split_words:"true"`
	AuthAudienceURL     string                 `default:"" split_words:"true"`
	DownloadRowLimit    int64                  `default:"10000" split_words:"true"`
	SafeSourceRefresh   bool                   `default:"false" split_words:"true"`
	ConnectionCacheSize int                    `default:"100" split_words:"true"`
	QueryCacheSizeBytes int64                  `default:"104857600" split_words:"true"` // 100MB by default
	// AllowHostAccess controls whether instance can use host credentials and
	// local_file sources can access directory outside repo
	AllowHostAccess bool `default:"false" split_words:"true"`
	// Redis server address host:port
	RedisURL string `default:"" split_words:"true"`
	// Sink type of activity client: noop (or empty string), kafka
	ActivitySinkType string `default:"" split_words:"true"`
	// Sink period of a buffered activity client in millis
	ActivitySinkPeriodMs int `default:"1000" split_words:"true"`
	// Max queue size of a buffered activity client
	ActivityMaxBufferSize int `default:"1000" split_words:"true"`
	// Kafka brokers of an activity client's sink
	ActivitySinkKafkaBrokers string `default:"" split_words:"true"`
	// Kafka topic of an activity client's sink
	ActivitySinkKafkaTopic string `default:"" split_words:"true"`
}

// StartCmd starts a stand-alone runtime server. It only allows configuration using environment variables.
func StartCmd(cliCfg *config.Config) *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start stand-alone runtime server",
		Run: func(cmd *cobra.Command, args []string) {
			// Load .env (note: fails silently if .env has errors)
			_ = godotenv.Load()

			// Init config
			var conf Config
			err := envconfig.Process("rill_runtime", &conf)
			if err != nil {
				fmt.Printf("failed to load config: %s\n", err.Error())
				os.Exit(1)
			}

			// Init logger
			cfg := zap.NewProductionConfig()
			cfg.Level.SetLevel(conf.LogLevel)
			cfg.EncoderConfig.NameKey = zapcore.OmitKey
			logger, err := cfg.Build()
			if err != nil {
				fmt.Printf("error: failed to create logger: %s\n", err.Error())
				os.Exit(1)
			}

			// Init telemetry
			shutdown, err := observability.Start(cmd.Context(), logger, &observability.Options{
				MetricsExporter: conf.MetricsExporter,
				TracesExporter:  conf.TracesExporter,
				ServiceName:     "runtime-server",
				ServiceVersion:  cliCfg.Version.String(),
			})
			if err != nil {
				logger.Fatal("error starting telemetry", zap.Error(err))
			}
			defer func() {
				// Allow 10 seconds to gracefully shutdown telemetry
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				err := shutdown(ctx)
				if err != nil {
					logger.Error("telemetry shutdown failed", zap.Error(err))
				}
			}()

			// Init activity client and sink to collect and emit activity events
			var sink activity.Sink
			switch conf.ActivitySinkType {
			case "", "noop":
				sink = activity.NewNoopSink()
			case "kafka":
				sink, err = activity.NewKafkaSink(conf.ActivitySinkKafkaBrokers, conf.ActivitySinkKafkaTopic)
				if err != nil {
					logger.Fatal("failed to create a kafka sink", zap.Error(err))
				}
			default:
				logger.Fatal(fmt.Sprintf("unknown activity sink type: %s", conf.ActivitySinkType))
			}
			activityOpts := activity.BufferedClientOptions{
				Sink:       sink,
				SinkPeriod: time.Duration(conf.ActivitySinkPeriodMs) * time.Millisecond,
				BufferSize: conf.ActivityMaxBufferSize,
				Logger:     logger,
			}
			activityClient := activity.NewBufferedClient(activityOpts)

			// Init runtime
			opts := &runtime.Options{
				ConnectionCacheSize: conf.ConnectionCacheSize,
				MetastoreConnector:  "metastore",
				QueryCacheSizeBytes: conf.QueryCacheSizeBytes,
				AllowHostAccess:     conf.AllowHostAccess,
				SafeSourceRefresh:   conf.SafeSourceRefresh,
				SystemConnectors: []*runtimev1.Connector{
					{
						Type:   conf.MetastoreDriver,
						Name:   "metastore",
						Config: map[string]string{"dsn": conf.MetastoreURL},
					},
				},
			}
			rt, err := runtime.New(opts, logger, activityClient)
			if err != nil {
				logger.Fatal("error: could not create runtime", zap.Error(err))
			}
			defer rt.Close()

			// Create ctx that cancels on termination signals
			ctx := graceful.WithCancelOnTerminate(context.Background())

			var limiter ratelimit.Limiter
			if conf.RedisURL == "" {
				limiter = ratelimit.NewNoop()
			} else {
				opts, err := redis.ParseURL(conf.RedisURL)
				if err != nil {
					logger.Fatal("failed to parse redis url", zap.Error(err))
				}
				limiter = ratelimit.NewRedis(redis.NewClient(opts))
			}

			// Init server
			srvOpts := &server.Options{
				HTTPPort:         conf.HTTPPort,
				GRPCPort:         conf.GRPCPort,
				AllowedOrigins:   conf.AllowedOrigins,
				ServePrometheus:  conf.MetricsExporter == observability.PrometheusExporter,
				AuthEnable:       conf.AuthEnable,
				AuthIssuerURL:    conf.AuthIssuerURL,
				AuthAudienceURL:  conf.AuthAudienceURL,
				DownloadRowLimit: &conf.DownloadRowLimit,
			}
			s, err := server.NewServer(ctx, srvOpts, rt, logger, limiter, activityClient)
			if err != nil {
				logger.Fatal("error: could not create server", zap.Error(err))
			}

			// Run server
			group, cctx := errgroup.WithContext(ctx)
			group.Go(func() error { return s.ServeGRPC(cctx) })
			group.Go(func() error { return s.ServeHTTP(cctx, nil) })
			err = group.Wait()
			if err != nil {
				logger.Error("server crashed", zap.Error(err))
				return
			}

			logger.Info("server shutdown gracefully")
		},
	}
	return startCmd
}
