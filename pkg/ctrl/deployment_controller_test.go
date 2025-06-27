package ctrl

import (
	// context "context"
	"context"
	"testing"
	"time"

	testutil "github.com/ibra86/k8s-controller-patterns/pkg/testutil"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func int32Ptr(i int32) *int32 { return &i }
func TestDeploymentReconciler_BasicFlow(t *testing.T) {
	mgr, k8sClient, _, cleanup := testutil.StartTestManager(t)
	defer cleanup()

	err := AddDeploymentController(mgr)
	require.NoError(t, err)

	go func() {
		_ = mgr.Start(context.Background())
	}()

	ctx := context.Background()
	name := "test-deployment"
	ns := "default"

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "nginx", Image: "nginx"},
					},
				},
			},
		},
	}
	if err := k8sClient.Create(ctx, dep); err != nil {
		t.Fatalf("Falied to create Deployment: %v", err)
	}

	time.Sleep(1 * time.Second) // wait to allow the reconciled to be triggered

	var got appsv1.Deployment
	err = k8sClient.Get(
		ctx,
		client.ObjectKey{Name: name, Namespace: ns},
		&got,
	)
	require.NoError(t, err)

}
