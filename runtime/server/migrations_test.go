package server

import (
	"context"
	"path/filepath"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	_ "github.com/rilldata/rill/runtime/services/catalog/artifacts/yaml"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/sources"
	"github.com/rilldata/rill/runtime/services/catalog/testutils"
	"github.com/stretchr/testify/require"
)

const TestDataPath = "../../web-local/test/data"

var AdBidsCsvPath = filepath.Join(TestDataPath, "AdBids.csv")
var AdImpressionsCsvPath = filepath.Join(TestDataPath, "AdImpressions.tsv")

const AdBidsRepoPath = "/sources/AdBids.yaml"
const AdBidsNewRepoPath = "/sources/AdBidsNew.yaml"
const AdBidsModelRepoPath = "/models/AdBids_model.sql"

var AdBidsAffectedPaths = []string{AdBidsRepoPath, AdBidsModelRepoPath}
var AdBidsNewAffectedPaths = []string{AdBidsNewRepoPath, AdBidsModelRepoPath}

func TestServer_PutFileAndReconcile(t *testing.T) {
	ctx := context.Background()
	server, instanceId := getTestServer(t)

	service, err := server.serviceCache.createCatalogService(ctx, server, instanceId)
	require.NoError(t, err)

	artifact := testutils.CreateSource(t, service, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
	resp, err := server.PutFileAndReconcile(ctx, &runtimev1.PutFileAndReconcileRequest{
		InstanceId: instanceId,
		Path:       AdBidsRepoPath,
		Blob:       artifact,
	})
	require.NoError(t, err)
	require.Len(t, resp.Errors, 0)
	require.Equal(t, []string{AdBidsRepoPath}, resp.AffectedPaths)
	testutils.AssertTable(t, service, "AdBids", AdBidsRepoPath)

	artifact = testutils.CreateModel(t, service, "AdBids_model", "select * from AdBids", AdBidsModelRepoPath)
	resp, err = server.PutFileAndReconcile(ctx, &runtimev1.PutFileAndReconcileRequest{
		InstanceId: instanceId,
		Path:       AdBidsModelRepoPath,
		Blob:       artifact,
	})
	require.NoError(t, err)
	require.Len(t, resp.Errors, 0)

	// replace with same name different file
	artifact = testutils.CreateSource(t, service, "AdBids", AdImpressionsCsvPath, AdBidsRepoPath)
	resp, err = server.PutFileAndReconcile(ctx, &runtimev1.PutFileAndReconcileRequest{
		InstanceId: instanceId,
		Path:       AdBidsRepoPath,
		Blob:       artifact,
	})
	require.NoError(t, err)
	require.Len(t, resp.Errors, 0)
	require.Equal(t, AdBidsAffectedPaths, resp.AffectedPaths)
	testutils.AssertTable(t, service, "AdBids", AdBidsRepoPath)

	// refresh the source without changes
	refreshResp, err := server.RefreshAndReconcile(ctx, &runtimev1.RefreshAndReconcileRequest{
		InstanceId: instanceId,
		Path:       AdBidsRepoPath,
	})
	require.NoError(t, err)
	require.Len(t, refreshResp.Errors, 0)
	require.Equal(t, AdBidsAffectedPaths, resp.AffectedPaths)

	// rename
	testutils.CreateSource(t, service, "AdBidsNew", AdBidsCsvPath, AdBidsRepoPath)
	testutils.CreateModel(t, service, "AdBids_model", "select * from AdBidsNew", AdBidsModelRepoPath)
	renameResp, err := server.RenameFileAndReconcile(ctx, &runtimev1.RenameFileAndReconcileRequest{
		InstanceId: instanceId,
		FromPath:   AdBidsRepoPath,
		ToPath:     AdBidsNewRepoPath,
	})
	require.NoError(t, err)
	require.Len(t, renameResp.Errors, 0)
	testutils.AssertTableAbsence(t, service, "AdBids")
	require.Equal(t, []string{AdBidsNewRepoPath, AdBidsModelRepoPath, AdBidsRepoPath}, renameResp.AffectedPaths)
	testutils.AssertTable(t, service, "AdBidsNew", AdBidsNewRepoPath)

	// delete
	delResp, err := server.DeleteFileAndReconcile(ctx, &runtimev1.DeleteFileAndReconcileRequest{
		InstanceId: instanceId,
		Path:       AdBidsNewRepoPath,
	})
	require.NoError(t, err)
	require.Len(t, delResp.Errors, 1)
	testutils.AssertTableAbsence(t, service, "AdBids")
	require.Equal(t, AdBidsNewAffectedPaths, delResp.AffectedPaths)
	testutils.AssertTableAbsence(t, service, "AdBidsNew")
}
