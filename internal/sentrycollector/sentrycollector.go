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

package sentrycollector

import (
	"strings"
	"sync"
	"time"

	"github.com/atlassian/go-sentry-api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var organisation sentry.Organization
var teams []sentry.Team
var projects []sentry.Project
var lastScan = make(map[string]int64)

// sentryCollector implements the prometheus.Collector interface
type sentryCollector struct {
	projectInfo   *prometheus.Desc
	projectErrors *prometheus.Desc
}

// NewSentryCollector returns an instance of the sentryCollector
func NewSentryCollector() *sentryCollector {
	return &sentryCollector{
		projectInfo: prometheus.NewDesc(
			"sentry_project_info",
			"Informational series so that Projects can be linked to Teams",
			[]string{"organisation", "team", "project"},
			nil,
		),
		projectErrors: prometheus.NewDesc(
			"sentry_project_errors",
			"Records the number of errors of a particular type for the specific project",
			[]string{"organisation", "project", "query"},
			nil,
		),
	}
}

// Describe adds the series to the collector
func (collector *sentryCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.projectInfo
	ch <- collector.projectErrors
}

// Collect handles incoming metric requests
func (collector *sentryCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now().Unix()
	log.Debug().Int64("now", start).Msg("Compiling metrics")

	// Create a Sentry client to query the API
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
	client, err := sentry.NewClient(
		viper.GetString("token"),
		url,
		timeout,
	)
	if err != nil {
		log.Error().Err(err).Msg("Could not initialise Sentry Client")
	}

	// get a slice of projects to include if specified
	var includeProjects []string
	if viper.IsSet("include_projects") {
		includeProjects = strings.Split(viper.GetString("include_projects"), ",")
	}

	// Compile the various metrics (if TTL hasn't expired)
	fetchOrganisation(*client)
	fetchTeams(*client)
	fetchProjects(*client)

	for _, team := range teams {
		for _, project := range *team.Projects {
			if len(includeProjects) == 0 || existsInSlice(*project.Slug, includeProjects) {
				ch <- prometheus.MustNewConstMetric(
					collector.projectInfo,
					prometheus.CounterValue,
					1,
					*organisation.Slug,
					*team.Slug,
					*project.Slug,
				)
			}
		}
	}

	if lastScan["errors"] == 0 {
		lastScan["errors"] = time.Now().Add(time.Second * -10).Unix()
	}
	var wg sync.WaitGroup
	for _, project := range projects {
		queries := []string{"received", "rejected", "blacklisted", "generated"}
		for _, query := range queries {
			wg.Add(1)
			go func(p sentry.Project, q string) {
				defer wg.Done()
				if len(includeProjects) == 0 || existsInSlice(*p.Slug, includeProjects) {
					count, err := fetchErrorCount(*client, p, q)
					if err != nil {
						log.Error().Err(err).Msg("Could not fetch project stats")
					} else {
						ch <- prometheus.MustNewConstMetric(
							collector.projectErrors,
							prometheus.GaugeValue,
							count,
							*organisation.Slug,
							*p.Slug,
							q,
						)
					}

				}
			}(project, query)
		}
	}
	wg.Wait()
	lastScan["errors"] = time.Now().Unix()
	end := time.Now().Unix()
	log.Debug().Int64("now", end).Int64("duration", end-start).Msg("Done compiling metrics")
}

// Check whether a specific string value exists in a string slice
func existsInSlice(value string, slice []string) bool {
	for _, s := range slice {
		if s == value {
			return true
		}
	}
	return false
}

// fetchOrganisation updates the global organisation object from data obtained from the
// Sentry API if the TTL has expired
func fetchOrganisation(client sentry.Client) {
	now := time.Now().Unix()
	if now-lastScan["organisation"] <= viper.GetInt64("ttl_organisation") {
		return
	}
	log.Info().Msg("Organisation TTL expired, refreshing")
	var err error
	organisation, err = client.GetOrganization(viper.GetString("organisation_name"))
	if err != nil {
		log.Error().
			Err(err).
			Str("organisation", viper.GetString("organisation_name")).
			Msg("Could not fetch organisation")
	}
	lastScan["organisation"] = time.Now().Unix()
}

// fetchTeams updates the global teams object from data obtained from the
// Sentry API if the TTL has expired
func fetchTeams(client sentry.Client) {
	now := time.Now().Unix()
	if now-lastScan["teams"] <= viper.GetInt64("ttl_teams") {
		return
	}
	log.Info().Msg("Teams TTL expired, refreshing")
	var err error
	teams, err = client.GetOrganizationTeams(organisation)
	if err != nil {
		log.Error().Err(err).Msg("Could not fetch organisation teams")
	}
	lastScan["teams"] = time.Now().Unix()
}

// fetchProjects updates the global projects object from data obtained from the
// Sentry API if the TTL has expired
func fetchProjects(client sentry.Client) {
	now := time.Now().Unix()
	if now-lastScan["projects"] <= viper.GetInt64("ttl_projects") {
		return
	}
	log.Info().Msg("Project TTL expired, refreshing")
	projects = []sentry.Project{}
	for ok := true; ok; {
		results, link, err := client.GetOrgProjects(organisation)
		if err != nil {
			log.Error().Err(err).Msg("Could not fetch organisation projects")
		}
		projects = append(projects, results...)
		if !link.Next.Results {
			break
		}
	}
	lastScan["projects"] = time.Now().Unix()
}

// fetchErrorCount queries the Sentry API for the error counts of the particular type
// for the specified project.  If there are multiple 10s buckets returned, it adds them
// together to return a single count.
func fetchErrorCount(client sentry.Client, project sentry.Project, query string) (float64, error) {
	resolution := "10s"
	var err error
	var c []sentry.Stat
	// Retry 3 times to fetch stats if there's a failure, with a 3s break between retries
	for i := 0; i < 3; i++ {
		c, err = client.GetProjectStats(
			organisation,
			project,
			sentry.StatQuery(query),
			lastScan["errors"],
			time.Now().Unix(),
			&resolution,
		)
		if err != nil {
			// sleep for 3 seconds and try again
			log.Debug().Msg("Could not fetch stats. Retrying")
			time.Sleep(time.Second * 3)
		} else {
			err = nil
			break
		}
	}
	if err != nil {
		return 0, err
	}
	var total float64
	for _, count := range c {
		total += count[1]
	}
	return total, nil
}
