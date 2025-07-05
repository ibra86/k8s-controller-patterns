package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/buaazp/fasthttprouter"
	"github.com/go-logr/zerologr"
	_ "github.com/ibra86/k8s-controller-patterns/docs"
	"github.com/ibra86/k8s-controller-patterns/pkg/api"
	"github.com/ibra86/k8s-controller-patterns/pkg/ctrl"
	"github.com/ibra86/k8s-controller-patterns/pkg/informer"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrlruntime "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	frontendv1alpha1 "github.com/ibra86/k8s-controller-patterns/pkg/apis/frontend/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

var serverPort int
var serverKubeconfig string
var serverInCluster bool
var enableLeaderElection bool
var leaderElectionNamespace string
var metricsPort int
var enableMCP bool
var mcpPort int
var FrontendAPI *api.FrontendPageAPI

func getServerKubeClient(kubeconfigPath string, inCluster bool) (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error
	if inCluster {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a FastHTTP server",
	Run: func(cmd *cobra.Command, args []string) {
		ConfigureLogger(logLevel)

		logf.SetLogger(zap.New(zap.UseDevMode(true)))
		logf.SetLogger(zerologr.New(&log.Logger))

		scheme := runtime.NewScheme()
		if err := clientgoscheme.AddToScheme(scheme); err != nil {
			log.Error().Err(err).Msg("Failed to add client-go scheme")
			os.Exit(1)
		}

		if err := frontendv1alpha1.AddToScheme(scheme); err != nil {
			log.Error().Err(err).Msg("Failed to add FrontendPage scheme")
			os.Exit(1)
		}

		clientset, err := getServerKubeClient(serverKubeconfig, serverInCluster)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create Kubernetes client")
			os.Exit(1)
		}
		ctx := context.Background()

		// Start controller-runtime manager and controller
		mgr, err := ctrlruntime.NewManager(
			ctrlruntime.GetConfigOrDie(),
			manager.Options{
				Scheme:                  scheme,
				LeaderElection:          enableLeaderElection,
				LeaderElectionID:        "k8s-controllers-leader-election",
				LeaderElectionNamespace: leaderElectionNamespace,
				Metrics:                 server.Options{BindAddress: fmt.Sprintf(":%d", metricsPort)},
			},
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create controller-runtime manager")
			os.Exit(1)
		}
		if err := ctrl.AddDeploymentController(mgr); err != nil {
			log.Error().Err(err).Msg("Failed to add deployment controller")
			os.Exit(1)
		}
		if err := ctrl.AddFrontendController(mgr); err != nil {
			log.Error().Err(err).Msg("Failed to add frontend controller")
			os.Exit(1)
		}

		go informer.StartDeploymentInformer(ctx, clientset)
		go func() {
			log.Info().Msg("Starting controller-runtime manager...")
			if err := mgr.Start(cmd.Context()); err != nil {
				log.Error().Err(err).Msg("Manager exited with error")
				os.Exit(1)
			}
		}()

		// API router
		router := fasthttprouter.New()
		frontendAPI := &api.FrontendPageAPI{
			K8sClient: mgr.GetClient(),
			Namespace: "default",
		}
		router.GET("/", func(ctx *fasthttp.RequestCtx) {
			_, _ = fmt.Fprintf(ctx, "hello from FastHTTP")
		})
		router.GET("/api/frontendpages", frontendAPI.ListFrontendPages)
		router.POST("/api/frontendpages", frontendAPI.CreateFrontendPage)
		router.GET("/api/frontendpages/:name", frontendAPI.GetFrontendPage)
		router.PUT("/api/frontendpages/:name", frontendAPI.UpdateFrontendPage)
		router.DELETE("/api/frontendpages/:name", frontendAPI.DeleteFrontendPage)

		router.GET(
			"/swagger/*any",
			fasthttpadaptor.NewFastHTTPHandler(
				httpSwagger.Handler(
					httpSwagger.URL("swagger/doc.json"),
				),
			),
		)

		// legacy endpoint for deployments
		handler := func(ctx *fasthttp.RequestCtx) {
			log.Info().Msg("Deployments request received")
			ctx.Response.Header.Set("Content-Type", "application/json")
			deployments := informer.GetDeploymentNames()
			log.Info().Msgf("Deployments: %v", deployments)
			ctx.SetStatusCode(200)
			_, _ = ctx.Write([]byte("["))
			for i, name := range deployments {
				_, _ = ctx.WriteString("\"")
				_, _ = ctx.WriteString(name)
				_, _ = ctx.WriteString("\"")
				if i < len(deployments)-1 {
					_, _ = ctx.WriteString(",")
				}
			}
			_, _ = ctx.Write([]byte("]"))
		}
		router.GET("/deployments", handler)

		if enableMCP {
			go func() {
				mcpServer := NewMCPServer("K8s Controller MCP", appVersion)
				sseServer := mcpserver.NewSSEServer(
					mcpServer,
					mcpserver.WithBaseURL(fmt.Sprintf("http://:%d", mcpPort)),
				)
				log.Info().Msgf("Starting MCP server in SSE mode on port %d", mcpPort)
				if err := sseServer.Start(fmt.Sprintf(":%d", mcpPort)); err != nil {
					log.Fatal().Err(err).Msg("MCP SSE server error")
				}
			}()
			log.Info().Msgf("MCP server is ready on port %d", mcpPort)

		}

		addr := fmt.Sprintf(":%d", serverPort)
		log.Info().Msgf("Starting FastHTTP server on %s (version: %s)", addr, appVersion)

		if err := fasthttp.ListenAndServe(addr, router.Handler); err != nil {
			log.Error().Err(err).Msg("Error starting FastHTTP server")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().IntVar(&serverPort, "port", 8080, "Port to run the server on")
	serverCmd.Flags().StringVar(&serverKubeconfig, "kubeconfig", "", "Path to the kubeconfig file")
	serverCmd.Flags().BoolVar(&serverInCluster, "in-cluster", false, "Use in-cluster Kubernetes config")
	serverCmd.Flags().BoolVar(&enableLeaderElection, "enable-leader-election", true, "Enable leader election for controller manager")
	serverCmd.Flags().StringVar(&leaderElectionNamespace, "leader-election-namespace", "default", "Namespace for leader election")
	serverCmd.Flags().IntVar(&metricsPort, "metrics-port", 8081, "Port for controller manager metrics")
	serverCmd.Flags().BoolVar(&enableMCP, "enable-mcp", false, "Enable MCP server")
	serverCmd.Flags().IntVar(&mcpPort, "mcp-port", 9090, "Port for MCP server")
}
