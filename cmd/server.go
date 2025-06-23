package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/valyala/fasthttp"
)

var serverPort int
var listenAndServe = fasthttp.ListenAndServe

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a FastHTTP server",
	Run: func(cmd *cobra.Command, args []string) {
		ConfigureLogger(logLevel)
		handler := func(ctx *fasthttp.RequestCtx) {
			log.Info().
				Str("method", string(ctx.Method())).
				Str("path", string(ctx.Path())).
				Str("remoteAddr", ctx.RemoteAddr().String()).
				Msg("Incoming HTTP request")

			fmt.Fprintf(ctx, "hello from FastHTTP")
		}

		addr := fmt.Sprintf(":%d", serverPort)
		log.Info().Msgf("Starting FastHTTP server on %s", addr)

		if err := listenAndServe(addr, handler); err != nil {
			log.Error().Err(err).Msg("Error starting FastHTTP server")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().IntVar(&serverPort, "port", 8080, "Port to run the server on")
}

func GetServerCmd() *cobra.Command {
	return serverCmd // needed to access it from test
}
