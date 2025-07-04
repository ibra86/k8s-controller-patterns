package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	frontendv1alpha1 "github.com/ibra86/k8s-controller-patterns/pkg/apis/frontend/v1alpha1"
	myctrl "github.com/ibra86/k8s-controller-patterns/pkg/ctrl"
	"github.com/ibra86/k8s-controller-patterns/pkg/testutil"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttprouter"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func setupTestAPIWithManager(t *testing.T) (*FrontendPageAPI, client.Client, func()) {
	mgr, k8sClient, _, cleanup := testutil.StartTestManager(t)
	require.NoError(t, myctrl.AddFrontendController(mgr))

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		_ = mgr.Start(ctx)
	}()

	if ok := mgr.GetCache().WaitForCacheSync(ctx); !ok {
		cancel()
		t.Fatal("cache did not sync")
	}

	api := &FrontendPageAPI{
		K8sClient: k8sClient,
		Namespace: "default",
	}

	return api, k8sClient, func() {
		cancel()
		cleanup()
	}
}

func cleanupFrontendPages(t *testing.T, c client.Client, ns string) {
	ctx := context.Background()
	var pages frontendv1alpha1.FrontendPageList
	require.NoError(t, c.List(ctx, &pages, client.InNamespace(ns)))
	for _, p := range pages.Items {
		require.NoError(t, c.Delete(ctx, &p))
	}
}

// adapter for func(*fasthttp.RequestCtx) to be used as fathttprouter.Handle
func adaptHandler(h func(ctx *fasthttp.RequestCtx)) fasthttprouter.Handle {
	return func(ctx *fasthttp.RequestCtx, _ fasthttprouter.Params) {
		h(ctx)
	}
}

func doRequest(
	router *fasthttprouter.Router,
	method,
	uri string,
	body []byte,
) *fasthttp.Response {
	ctx := &fasthttp.RequestCtx{}
	req := &ctx.Request
	resp := &ctx.Response
	ctx.Init(req, nil, nil)
	req.Header.SetMethod(method)
	req.SetRequestURI(uri)
	if body != nil {
		req.SetBody(body)
	}

	if (method == http.MethodGet ||
		method == http.MethodPut ||
		method == http.MethodDelete) &&
		strings.HasPrefix(uri, "/api/frontendpages/") {
		parts := strings.Split(uri, "/")
		if len(parts) > 3 {
			ctx.SetUserValue("name", parts[3])
		}
	}
	router.Handler(ctx)
	return resp
}

func getDeployment(
	t *testing.T,
	c client.Client,
	name,
	ns string,
	timeout time.Duration) *appsv1.Deployment {
	var dep appsv1.Deployment
	var lastErr error
	end := time.Now().Add(timeout)
	for time.Now().Before(end) {
		t.Logf("Checking for deployment %s/%s", ns, name)
		err := c.Get(context.Background(), client.ObjectKey{Name: name, Namespace: ns}, &dep)
		if err == nil {
			return &dep
		}
	}
	t.Fatalf("Deployment %s/%s not found after %v: %v", ns, name, timeout, lastErr)
	return nil
}

func TestFrontendPageAPI_E2E(t *testing.T) {
	log.SetLogger(zap.New(zap.UseDevMode(true)))

	id := uuid.NewString()[:8]
	resourceName := "test-frontend-page-" + id

	api, k8sClient, cleanup := setupTestAPIWithManager(t)
	defer cleanup()

	cleanupFrontendPages(t, k8sClient, "default")

	router := fasthttprouter.New()
	router.GET("/api/frontendpages", adaptHandler(api.ListFrontendPages))
	router.GET("/api/frontendpages/:name", adaptHandler(api.GetFrontendPage))
	router.POST("/api/frontendpages", adaptHandler(api.CreateFrontendPage))
	router.PUT("/api/frontendpages/:name", adaptHandler(api.UpdateFrontendPage))
	router.DELETE("/api/frontendpages/:name", adaptHandler(api.DeleteFrontendPage))

	t.Logf("[TEST] POST /api/frontendpages (name=%s)", resourceName)
	createDoc := FrontendPageDoc{
		Name:     resourceName,
		Contents: "<h1>Hello</h1>",
		Image:    "nginx:latest",
		Replicas: 2,
	}
	body, _ := json.Marshal(&frontendv1alpha1.FrontendPage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      createDoc.Name,
			Namespace: "default",
		},
		Spec: frontendv1alpha1.FrontendPageSpec{
			Contents: createDoc.Contents,
			Image:    createDoc.Image,
			Replicas: createDoc.Replicas,
		},
		TypeMeta: frontendv1alpha1.FrontendPage{}.TypeMeta,
	})
	resp := doRequest(router, http.MethodPost, "/api/frontendpages", body)
	t.Logf("Create response body: %s", resp.Body())
	require.Equal(t, http.StatusCreated, resp.StatusCode())

	// create deployment
	dep := getDeployment(t, k8sClient, resourceName, "default", 2*time.Second)
	t.Logf("Deployment after create: name=%s replicas=%v image=%s", dep.Name, *dep.Spec.Replicas, dep.Spec.Template.Spec.Containers[0].Image)

	t.Logf("[TEST] PUT /api/frontendpages/%s", resourceName)
	var existing frontendv1alpha1.FrontendPage
	require.NoError(t,
		k8sClient.Get(context.Background(),
			client.ObjectKey{
				Name:      resourceName,
				Namespace: "default",
			}, &existing))
	updateDoc := createDoc
	updateDoc.Contents = "<h1>Updated</h1>"
	updateDoc.Replicas = 1
	body, _ = json.Marshal(&frontendv1alpha1.FrontendPage{
		ObjectMeta: metav1.ObjectMeta{
			Name:            updateDoc.Name,
			Namespace:       "default",
			ResourceVersion: existing.ResourceVersion,
		},
		Spec: frontendv1alpha1.FrontendPageSpec{
			Contents: updateDoc.Contents,
			Image:    updateDoc.Image,
			Replicas: updateDoc.Replicas,
		},
		TypeMeta: frontendv1alpha1.FrontendPage{}.TypeMeta,
	})
	resp = doRequest(router, http.MethodPut, "/api/frontendpages/"+resourceName, body)
	t.Logf("Update response body: %s", resp.Body())
	require.Equal(t, http.StatusOK, resp.StatusCode())

	// new deployment
	dep = getDeployment(t, k8sClient, resourceName, "default", 5*time.Second)
	t.Logf("Deployment after update: name=%s replicas=%v image=%s", dep.Name, *dep.Spec.Replicas, dep.Spec.Template.Spec.Containers[0].Image)

	t.Logf("[TEST] DELETE /api/frontendpages/%s", resourceName)
	resp = doRequest(router, http.MethodDelete, "/api/frontendpages/"+resourceName, nil)
	require.Equal(t, http.StatusNoContent, resp.StatusCode())

	end := time.Now().Add(2 * time.Second) // wait for deployment to be deleted
	for time.Now().Before(end) {
		var dep appsv1.Deployment
		err := k8sClient.Get(
			context.Background(),
			client.ObjectKey{Name: resourceName, Namespace: "default"},
			&dep,
		)
		if err != nil {
			t.Logf("Deployment deleted as expected")
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

}
