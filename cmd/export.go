package cmd

import (
	"github.com/MaggotMouth/sentry-exporter/internal/sentry"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export stats",
	Long:  `Export stats from Sentry to Prometheus Push Gateway`,
	Run: func(cmd *cobra.Command, args []string) {
		exportStats()
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringP("organisation", "o", "", "Organisation to scan for teams/projects")
	exportCmd.Flags().StringP("team", "t", "", "Team to scan for projects")
	exportCmd.Flags().StringP("project", "p", "", "Project slug to get stats for")
}

func exportStats() {
	var url *string
	if viper.IsSet("api_url") {
		uVal := viper.GetString("api_url")
		url = &uVal
	}
	var timeout *int
	if viper.IsSet("timeout") {
		tVal := viper.GetInt("timeout")
		timeout = &tVal
	}
	c := sentry.NewSentryClient(
		viper.GetString("token"),
		url,
		timeout,
	)
	p := sentry.NewSentryProcessor(c)
	s := p.GetStats(viper.GetString("organisation_name"))
	log.Info().Interface("stats", s).Msg("Huzzah")
	log.Info().Str("stats", s.String()).Msg("Huzzah")
}
