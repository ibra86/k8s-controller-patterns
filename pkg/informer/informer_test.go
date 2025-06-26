package informer

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	testutil "github.com/ibra86/k8s-controller-patterns/pkg/testutil"
)

func TestGetDeploymentName(t *testing.T) {
	dep := &metav1.PartialObjectMetadata{}
	dep.SetName("test-deployment")
	name := getDeploymentName(dep)
	if name != "test-deployment" {
		t.Errorf("expected 'test-deployment', got %q", name)
	}
	name = getDeploymentName("not-an-object")
	if name != "unknown" {
		t.Errorf("expected 'unknown', got %q", name)
	}

}

func TestStartDeploymentInformer_CoversFunction(t *testing.T) {
	_, clientset, cleanup := testutil.SetupEnv(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		StartDeploymentInformer(ctx, clientset)
	}()

	time.Sleep(1 * time.Second)
	cancel()
}

func TestStartDeploymentInformer(t *testing.T) {
	_, clientset, cleanup := testutil.SetupEnv(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	added := make(chan string, 2) // patch log to write to a buffer

	factory := informers.NewSharedInformerFactoryWithOptions(
		clientset,
		30*time.Second,
		informers.WithNamespace("default"),
	)
	informer := factory.Apps().V1().Deployments().Informer()
	_, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			if d, ok := obj.(metav1.Object); ok {
				added <- d.GetName()
			}
		},
	})
	if err != nil {
		log.Error().Msgf("failed to register event handler: %v", err)
	}

	go func() {
		defer wg.Done()
		factory.Start(ctx.Done())
		factory.WaitForCacheSync(ctx.Done())
		<-ctx.Done()
	}()

	found := map[string]bool{} // wait for events
	for range 2 {
		select {
		case name := <-added:
			found[name] = true
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for deployment add events")
		}
	}

	require.True(t, found["sample-deployment-1"])
	require.True(t, found["sample-deployment-2"])

	cancel()
	wg.Wait()

}
