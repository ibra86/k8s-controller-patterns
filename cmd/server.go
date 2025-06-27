package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/ibra86/k8s-controller-patterns/pkg/ctrl"
	"github.com/ibra86/k8s-controller-patterns/pkg/informer"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/valyala/fasthttp"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrlruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var serverPort int
var serverKubeconfig string
var serverInCluster bool

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

		clientset, err := getServerKubeClient(serverKubeconfig, serverInCluster)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create Kubernetes client")
			os.Exit(1)
		}
		ctx := context.Background()
		go informer.StartDeploymentInformer(ctx, clientset)

		// Start controller-runtime manager and controller
		mgr, err := ctrlruntime.NewManager(ctrlruntime.GetConfigOrDie(), manager.Options{})
		if err != nil {
			log.Error().Err(err).Msg("Failed to create controller-runtime manager")
			os.Exit(1)
		}
		if err := ctrl.AddDeploymentController(mgr); err != nil {
			log.Error().Err(err).Msg("Failed to add deployment controller")
			os.Exit(1)
		}
		go func() {
			log.Info().Msg("Starting controller-runtime manager...")
			if err := mgr.Start(cmd.Context()); err != nil {
				log.Error().Err(err).Msg("Failed to add deployment controller")
				os.Exit(1)
			}

		}()

		handler := func(ctx *fasthttp.RequestCtx) {
			log.Info().
				Str("method", string(ctx.Method())).
				Str("path", string(ctx.Path())).
				Str("remoteAddr", ctx.RemoteAddr().String()).
				Msg("Incoming HTTP request")

			requestID := uuid.New().String()
			ctx.Response.Header.Set("X-Request-ID", requestID)
			logger := log.With().Str("request_id", requestID).Logger()
			switch string(ctx.Path()) {
			case "/deployments":
				logger.Info().Msg("Deployments request received")
				ctx.Response.Header.Set("Content-Type", "application/json")
				deployments := informer.GetDeploymentNames()
				logger.Info().Msgf("Deployments: %v", deployments)
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
			default:
				logger.Info().Msg("Default request received")
				if _, err := fmt.Fprintf(ctx, "hello from FastHTTP"); err != nil {
					log.Printf("failed to write response: %v", err)
				}
			}

		}

		addr := fmt.Sprintf(":%d", serverPort)
		log.Info().Msgf("Starting FastHTTP server on %s (version: %s)", addr, appVersion)

		if err := fasthttp.ListenAndServe(addr, handler); err != nil {
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
}
