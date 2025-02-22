package reconcilers

import (
	"context"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	compilerv1 "github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindMigration, newMigrationReconciler)
}

type MigrationReconciler struct {
	C *runtime.Controller
}

func newMigrationReconciler(c *runtime.Controller) runtime.Reconciler {
	return &MigrationReconciler{C: c}
}

func (r *MigrationReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *MigrationReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	mig := self.GetMigration()

	from := mig.State.Version
	to := mig.Spec.Version

	if to-from > 100 {
		return runtime.ReconcileResult{Err: fmt.Errorf("difference between migration versions %d and %d is too large", from, to)}
	}

	for v := from; v <= to; v++ {
		err := r.executeMigration(ctx, self, v)
		if err != nil {
			if v != 0 {
				err = fmt.Errorf("failed to execute version %d: %w", v, err)
			}
			return runtime.ReconcileResult{Err: err}
		}

		mig.State.Version = v
		err = r.C.UpdateState(ctx, self.Meta.Name, self)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	return runtime.ReconcileResult{}
}

func (r *MigrationReconciler) executeMigration(ctx context.Context, self *runtimev1.Resource, version uint32) error {
	inst, err := r.C.Runtime.FindInstance(ctx, r.C.InstanceID)
	if err != nil {
		return err
	}

	spec := self.Resource.(*runtimev1.Resource_Migration).Migration.Spec
	state := self.Resource.(*runtimev1.Resource_Migration).Migration.State

	sql, err := compilerv1.ResolveTemplate(spec.Sql, compilerv1.TemplateData{
		Claims:    map[string]interface{}{},
		Variables: inst.ResolveVariables(),
		ExtraProps: map[string]interface{}{
			"version": version,
		},
		Self: compilerv1.TemplateResource{
			Meta:  self.Meta,
			Spec:  spec,
			State: state,
		},
		Resolve: func(ref compilerv1.ResourceName) (string, error) {
			return safeSQLName(ref.Name), nil
		},
		Lookup: func(name compilerv1.ResourceName) (compilerv1.TemplateResource, error) {
			if name.Kind == compilerv1.ResourceKindUnspecified {
				return compilerv1.TemplateResource{}, fmt.Errorf("can't resolve name %q without kind specified", name.Name)
			}
			res, err := r.C.Get(ctx, resourceNameFromCompiler(name))
			if err != nil {
				return compilerv1.TemplateResource{}, err
			}
			return compilerv1.TemplateResource{
				Meta:  res.Meta,
				Spec:  res.Resource.(*runtimev1.Resource_Model).Model.Spec,
				State: res.Resource.(*runtimev1.Resource_Model).Model.State,
			}, nil
		},
	})
	if err != nil {
		return fmt.Errorf("failed to resolve template: %w", err)
	}

	olap, release, err := r.C.AcquireOLAP(ctx, spec.Connector)
	if err != nil {
		return err
	}
	defer release()

	return olap.Exec(ctx, &drivers.Statement{
		Query:    sql,
		Priority: 100,
	})
}
