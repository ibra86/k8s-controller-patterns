package testutil

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	frontendv1alpha1 "github.com/ibra86/k8s-controller-patterns/pkg/apis/frontend/v1alpha1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func int32Ptr(i int32) *int32 { return &i }

func SetupEnv(t *testing.T) (*envtest.Environment, *kubernetes.Clientset, func()) {
	t.Helper()
	ctx := context.Background()
	env := &envtest.Environment{}

	cfg, err := env.Start()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	kubeconfig := clientcmdapi.NewConfig()
	kubeconfig.Clusters["envtest"] = &clientcmdapi.Cluster{
		Server:                   cfg.Host,
		CertificateAuthorityData: cfg.CAData,
	}
	kubeconfig.AuthInfos["envtest-user"] = &clientcmdapi.AuthInfo{
		ClientCertificateData: cfg.CertData,
		ClientKeyData:         cfg.KeyData,
	}
	kubeconfig.Contexts["envtest-context"] = &clientcmdapi.Context{
		Cluster:  "envtest",
		AuthInfo: "envtest-user",
	}
	kubeconfig.CurrentContext = "envtest-context"

	kubeconfigBytes, err := clientcmd.Write(*kubeconfig)
	require.NoError(t, err)
	err = os.WriteFile("/tmp/envtest.kubeconfig", kubeconfigBytes, 0644)
	require.NoError(t, err)

	clientset, err := kubernetes.NewForConfig(cfg)
	require.NoError(t, err)

	for i := 1; i <= 2; i++ {
		dep := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("sample-deployment-%d", i),
				Namespace: "default",
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
		_, err := clientset.
			AppsV1().
			Deployments("default").
			Create(ctx, dep, metav1.CreateOptions{})
		require.NoError(t, err)
	}
	cleanup := func() {
		_ = env.Stop()
	}
	return env, clientset, cleanup
}

func StartTestManager(t *testing.T) (
	mgr manager.Manager,
	k8sClient client.Client,
	restCfg *rest.Config,
	cleanup func(),
) {
	t.Helper()
	testScheme := runtime.NewScheme()

	require.NoError(t, scheme.AddToScheme(testScheme))
	require.NoError(t, frontendv1alpha1.AddToScheme(testScheme))
	metav1.AddToGroupVersion(testScheme, frontendv1alpha1.SchemeGroupVersion)
	require.NoError(t, apiextensionsv1.AddToScheme(testScheme))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	env := &envtest.Environment{
		CRDDirectoryPaths:        []string{"../../config/crd"},
		ErrorIfCRDPathMissing:    true,
		AttachControlPlaneOutput: false,
	}
	var startErr = make(chan error)
	var cfg *rest.Config
	var err error

	go func() {
		cfg, err = env.Start()
		startErr <- err
	}()

	select {
	case err := <-startErr:
		require.NoError(t, err, "Failed to start test environemt")
	case <-ctx.Done():
		t.Fatal("timeout waiting for test environment to start")
	}

	require.NotNil(t, cfg)

	mgr, err = manager.New(
		cfg,
		manager.Options{Scheme: testScheme, LeaderElection: false},
	)

	ctx, cancel = context.WithCancel(context.Background())
	go func() {
		_ = mgr.Start(ctx)
	}()

	k8sClient = mgr.GetClient()

	cleanup = func() {
		cancel()
		_ = env.Stop()
	}

	return mgr, k8sClient, cfg, cleanup
}
