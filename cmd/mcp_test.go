package cmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ibra86/k8s-controller-patterns/pkg/api"
	frontendv1alpha1 "github.com/ibra86/k8s-controller-patterns/pkg/apis/frontend/v1alpha1"
	"github.com/ibra86/k8s-controller-patterns/pkg/ctrl"
	"github.com/ibra86/k8s-controller-patterns/pkg/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func setupTestAPIWithManager(t *testing.T) (*api.FrontendPageAPI, client.Client, func()) {
	mgr, k8sClient, _, cleanup := testutil.StartTestManager(t)
	require.NoError(t, ctrl.AddFrontendController(mgr))
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		_ = mgr.Start(ctx)
	}()

	if ok:= mgr.GetCache().WaitForCacheSync(ctx); !ok {
		cancel()
		t.Fatal("cache did not sync")
	}

	apiInst := &api.FrontendPageAPI{
		K8sClient: k8sClient,
		Namespace: "default",
	}
	return apiInst, k8sClient, func(){ 
		cancel() 
		cleanup()
	}

}

func TestMCP_ListFrontendPagesHandler(t *testing.T){
	apiInst, k8sClient, cleanup := setupTestAPIWithManager(t)
	defer cleanup()
	api.FrontendAPI = apiInst

	// create frontend resources
	page1 := &frontendv1alpha1.FrontendPage{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mcp-page1",
			Namespace: "default",
		},
		Spec: frontendv1alpha1.FrontendPageSpec{
			Contents: "<h1>MCP Page-1</h1>",
			Image: "nginx:1.21",
			Replicas: 1,
		},
	}
	page2 := &frontendv1alpha1.FrontendPage{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mcp-page2",
			Namespace: "default",
		},
		Spec: frontendv1alpha1.FrontendPageSpec{
			Contents: "<h1>MCP Page-2</h1>",
			Image: "nginx:1.22",
			Replicas: 1,
		},
	}
	
	require.NoError(t, k8sClient.Create(context.Background(), page1))
	require.NoError(t, k8sClient.Create(context.Background(), page2))

	docs, err := api.FrontendAPI.ListFrontendPagesRaw(context.Background())
	require.NoError(t, err)
	require.Len(t, docs, 2)
	names := []string{docs[0].Name, docs[1].Name}
	require.Contains(t, names, "mcp-page1")
	require.Contains(t, names, "mcp-page2")
}
