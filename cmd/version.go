package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints version information of the application",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Msg("Version: " + version)
		log.Info().Msg("Build date: " + buildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
