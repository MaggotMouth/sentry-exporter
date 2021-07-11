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
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile string
var logLevel string
var logFormat string
var token string
var org string

const version = "0.2.0"
const buildDate = "2021/07/11 15:11"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sentry-exporter",
	Short: "Export your Sentry metrics to Prometheus",
	Long: `This application queries the Sentry API and exports the events counts to
a Prometheus Push Gateway.

It breaks down the stats per organisation, project, team and stat`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(
		&cfgFile,
		"config",
		"",
		"config file (default is $CURRENT_DIR/.sentry-exporter.yaml)",
	)
	rootCmd.PersistentFlags().StringVar(
		&logLevel,
		"loglevel",
		"info",
		"set the log level. options: trace|debug|info|warn|error|fatal|panic",
	)
	rootCmd.PersistentFlags().StringVar(
		&logFormat,
		"logformat",
		"text",
		"set the log format. options: text|json",
	)
	rootCmd.PersistentFlags().StringVar(
		&token,
		"token",
		"",
		"Sentry token",
	)
	rootCmd.PersistentFlags().StringVar(
		&org,
		"organisation",
		"",
		"Sentry organisation to query for statistics",
	)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find curr directory.
		curr, err := os.Getwd()
		cobra.CheckErr(err)

		// Search config in home directory with name ".sentry-exporter" (without extension).
		viper.AddConfigPath(curr)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".sentry-exporter")
	}
	viper.SetEnvPrefix("SENTRY_EXPORTER")

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	switch logLevel {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	}

	if logFormat == "text" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Default values for configuration
	viper.SetDefault("listen_address", ":9142")
	viper.SetDefault("ttl_organisation", 86400)
	viper.SetDefault("ttl_projects", 600)
	viper.SetDefault("ttl_teams", 3600)

	if token != "" {
		viper.SetDefault("token", token)
	}

	if org != "" {
		viper.SetDefault("organisation_name", org)
	}
}
