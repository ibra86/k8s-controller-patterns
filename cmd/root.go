/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func parseLogLevel(lvl string) zerolog.Level {
	switch strings.ToLower(lvl) {
	case "info":
		return zerolog.InfoLevel
	case "debug":
		return zerolog.DebugLevel
	case "trace":
		return zerolog.TraceLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

func ConfigureLogger(levelStr string) {
	level := parseLogLevel(levelStr)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(level)

	//nolint:staticcheck // QF1003: prefer if-else for
	if level == zerolog.TraceLevel {
		zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
			return fmt.Sprintf("%s:%d", file, line)
		}
		zerolog.CallerFieldName = "caller"
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "2006-01-02 15:04:05.000",
			PartsOrder: []string{
				zerolog.TimestampFieldName,
				zerolog.LevelFieldName,
				zerolog.CallerFieldName,
				zerolog.MessageFieldName,
			},
		}).With().Caller().Logger()
	} else if level == zerolog.DebugLevel {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "2006-01-02 15:04:05.000",
			PartsOrder: []string{
				zerolog.TimestampFieldName,
				zerolog.LevelFieldName,
				zerolog.MessageFieldName,
			},
		})
	} else {
		log.Logger = log.Output(os.Stderr)
	}
}

var logLevel string
var appVersion = "dev"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "k8s-controller-patterns",
	Short: "app version: " + appVersion,
	Run: func(cmd *cobra.Command, args []string) {
		ConfigureLogger(logLevel)

		log.Info().Msg("info log")
		log.Debug().Msg("info log")
		log.Trace().Msg("info log")
		log.Warn().Msg("info log")
		log.Error().Msg("info log")
		fmt.Println("k8s-controller cli with logging")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set log level: info, trace, debug, warn, error")
}
