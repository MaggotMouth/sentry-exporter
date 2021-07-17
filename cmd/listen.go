/*
   Copyright 2021 Willem Potgieter

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package cmd

import (
	"net/http"

	"github.com/MaggotMouth/sentry-exporter/internal/sentrycollector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var projectIncludes string

// listenCmd represents the listen command
var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Listen starts a web server",
	Long: `Listen starts a web server that exposes a /metrics endpoint to scrape with Prometheus.

This endpoint exposes metrics from the Sentry API`,
	Run: func(cmd *cobra.Command, args []string) {
		startListener()
	},
}

func init() {
	rootCmd.AddCommand(listenCmd)

	rootCmd.PersistentFlags().StringVar(
		&projectIncludes,
		"include-projects",
		"",
		"projects to include in the export (default include all projects)",
	)

	collector := sentrycollector.NewSentryCollector()
	prometheus.MustRegister(collector)
}

func startListener() {
	address := viper.GetString("listen_address")

	if projectIncludes != "" {
		viper.SetDefault("include_projects", projectIncludes)
	}

	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	))
	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Encountered an error")
	}
}
