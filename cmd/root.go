/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "k8s-controller-patterns",
	Short: "logging",
	Run: func(cmd *cobra.Command, args []string) {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		log.Info().Msg("info log")
		log.Debug().Msg("info log")
		log.Trace().Msg("info log")
		log.Warn().Msg("info log")
		log.Error().Msg("info log")
		fmt.Println("k8s-controller cli with logging")
	},
}

func Execute() {
	rootCmd.Execute()
}
