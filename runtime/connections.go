package runtime

import (
	"context"
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog"
)

func (r *Runtime) newMetaStore(ctx context.Context, instanceID string) (drivers.Handle, func(), error) {
	c, err := r.ConnectorDefByName(r.opts.MetastoreDriver)
	if err != nil {
		panic(err)
	}

	return r.connCache.get(ctx, instanceID, c.Type, r.variables(r.opts.MetastoreDriver, c.Configs, nil), true)
}

func (r *Runtime) Registry() drivers.RegistryStore {
	registry, ok := r.metastore.AsRegistry()
	if !ok {
		// Verified as registry in New, so this should never happen
		panic("metastore is not a registry")
	}
	return registry
}

func (r *Runtime) AcquireHandle(ctx context.Context, instanceID, connector string) (drivers.Handle, func(), error) {
	instance, err := r.FindInstance(ctx, instanceID)
	if err != nil {
		return nil, nil, err
	}

	if instance.RillYAML != nil {
		// defined in rill.yaml
		for _, c := range instance.RillYAML.Connectors {
			if c.Name == connector {
				return r.connCache.get(ctx, instanceID, c.Type, r.variables(connector, c.Configs, instance.ResolveVariables()), false)
			}
		}
	}
	if c, err := r.ConnectorDefByName(connector); err == nil { // connector found
		// defined in runtime options
		return r.connCache.get(ctx, instanceID, c.Type, r.variables(connector, c.Configs, instance.ResolveVariables()), true)
	}
	// neither defined in rill.yaml nor in runtime options, directly used in source
	return r.connCache.get(ctx, instanceID, connector, r.variables(connector, nil, instance.ResolveVariables()), false)
}

func (r *Runtime) Repo(ctx context.Context, instanceID string) (drivers.RepoStore, func(), error) {
	inst, err := r.FindInstance(ctx, instanceID)
	if err != nil {
		return nil, nil, err
	}

	c, shared, err := r.RepoDef(inst)
	if err != nil {
		return nil, nil, err
	}
	conn, release, err := r.connCache.get(ctx, instanceID, c.Type, r.variables(inst.RepoDriver, c.Configs, inst.ResolveVariables()), shared)
	if err != nil {
		return nil, nil, err
	}

	repo, ok := conn.AsRepoStore(instanceID)
	if !ok {
		release()
		// Verified as repo when instance is created, so this should never happen
		return nil, release, fmt.Errorf("connection for instance '%s' is not a repo", instanceID)
	}

	return repo, release, nil
}

func (r *Runtime) OLAP(ctx context.Context, instanceID string) (drivers.OLAPStore, func(), error) {
	inst, err := r.FindInstance(ctx, instanceID)
	if err != nil {
		return nil, nil, err
	}

	c, shared, err := r.OLAPDef(inst)
	if err != nil {
		return nil, nil, err
	}
	conn, release, err := r.connCache.get(ctx, instanceID, c.Type, r.variables(inst.OLAPDriver, c.Configs, inst.ResolveVariables()), shared)
	if err != nil {
		return nil, nil, err
	}

	olap, ok := conn.AsOLAP(instanceID)
	if !ok {
		release()
		// Verified as OLAP when instance is created, so this should never happen
		return nil, nil, fmt.Errorf("connection for instance '%s' is not an olap", instanceID)
	}

	return olap, release, nil
}

func (r *Runtime) Catalog(ctx context.Context, instanceID string) (drivers.CatalogStore, func(), error) {
	inst, err := r.FindInstance(ctx, instanceID)
	if err != nil {
		return nil, nil, err
	}

	if inst.EmbedCatalog {
		c, shared, err := r.OLAPDef(inst)
		if err != nil {
			return nil, nil, err
		}
		conn, release, err := r.connCache.get(ctx, instanceID, c.Type, r.variables(inst.OLAPDriver, c.Configs, inst.ResolveVariables()), shared)
		if err != nil {
			return nil, nil, err
		}

		store, ok := conn.AsCatalogStore(instanceID)
		if !ok {
			release()
			// Verified as CatalogStore when instance is created, so this should never happen
			return nil, nil, fmt.Errorf("instance cannot embed catalog")
		}

		return store, release, nil
	}

	store, ok := r.metastore.AsCatalogStore(instanceID)
	if !ok {
		return nil, nil, fmt.Errorf("metastore cannot serve as catalog")
	}
	return store, func() {}, nil
}

func (r *Runtime) NewCatalogService(ctx context.Context, instanceID string) (*catalog.Service, error) {
	// get all stores
	olapStore, releaseOLAP, err := r.OLAP(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	catalogStore, releaseCatalog, err := r.Catalog(ctx, instanceID)
	if err != nil {
		releaseOLAP()
		return nil, err
	}

	repoStore, releaseRepo, err := r.Repo(ctx, instanceID)
	if err != nil {
		releaseOLAP()
		releaseCatalog()
		return nil, err
	}

	registry := r.Registry()

	migrationMetadata := r.migrationMetaCache.get(instanceID)
	releaseFunc := func() {
		releaseOLAP()
		releaseCatalog()
		releaseRepo()
	}
	return catalog.NewService(catalogStore, repoStore, olapStore, registry, instanceID, r.logger, migrationMetadata, releaseFunc), nil
}

// TODO :: these can also be generated during reconcile itself ?
func (r *Runtime) variables(name string, def, variables map[string]string) map[string]any {
	vars := make(map[string]any, 0)
	for key, value := range def {
		vars[strings.ToLower(key)] = value
	}

	// connector variables are of format connector.name.var
	// there could also be other variables like allow_host_access, region etc which are global for all connectors
	prefix := fmt.Sprintf("connector.%s.", name)
	for key, value := range variables {
		if !strings.HasPrefix(key, "connector.") { // global variable
			vars[strings.ToLower(key)] = value
		} else if after, found := strings.CutPrefix(key, prefix); found { // connector specific variable
			vars[strings.ToLower(after)] = value
		}
	}
	vars["allow_host_access"] = r.opts.AllowHostAccess
	return vars
}

func (r *Runtime) ConnectorDefByName(name string) (*Connector, error) {
	for _, c := range r.opts.GlobalDrivers {
		if c.Name == name {
			return c, nil
		}
	}
	return nil, fmt.Errorf("connector %s doesn't exist", name)
}
