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

var (
	organisation        sentry.Organization
	teams               []sentry.Team
	projects            []sentry.Project
	lastScan            = make(map[string]int64)
	sentryClient        *sentry.Client
	includeProjects     []string
	includeTeams        []string
	apiSuccessCallCount float64
	apiFailureCallCount float64
)

const intialiseSentryError = "Could not initialise Sentry client"

// sentryCollector implements the prometheus.Collector interface
type sentryCollector struct {
	projectInfo   *prometheus.Desc
	projectErrors *prometheus.Desc
	apiCalls      *prometheus.Desc
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
		apiCalls: prometheus.NewDesc(
			"sentry_api_calls",
			"Records the number of calls made from the exporter to the Sentry API",
			[]string{"status"},
			nil,
		),
	}
}

// GetSentryClient returns the instantiated sentry client if it exists, or creates
// one for use throughout the package
func GetSentryClient() (*sentry.Client, error) {
	if sentryClient != nil {
		return sentryClient, nil
	}

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
		sentryClient = client
	}
	return client, err
}

// Describe adds the series to the collector
func (collector *sentryCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.projectInfo
	ch <- collector.projectErrors
	ch <- collector.apiCalls
}

// Collect handles incoming metric requests
func (collector *sentryCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now().Unix()
	apiSuccessCallCount = 0
	apiFailureCallCount = 0
	log.Debug().Int64("now", start).Msg("Compiling metrics")

	// get a slice of projects to include if specified
	if viper.IsSet("include_projects") {
		includeProjects = strings.Split(viper.GetString("include_projects"), ",")
	}

	// get a slice of teams to include if specified
	if viper.IsSet("include_teams") {
		includeTeams = strings.Split(viper.GetString("include_teams"), ",")
	}

	// Compile the various metrics (if TTL hasn't expired)
	fetchOrganisation()
	fetchTeams()
	fetchProjects()

	exportTeams(collector, ch)
	exportProjects(collector, ch)

	ch <- prometheus.MustNewConstMetric(
		collector.apiCalls,
		prometheus.GaugeValue,
		apiSuccessCallCount,
		"success",
	)
	ch <- prometheus.MustNewConstMetric(
		collector.apiCalls,
		prometheus.GaugeValue,
		apiFailureCallCount,
		"failure",
	)

	end := time.Now().Unix()
	log.Debug().Int64("now", end).Int64("duration", end-start).Msg("Done compiling metrics")
}

// exportTeams adds the sentry_project_info metric to the collector for export
func exportTeams(
	collector *sentryCollector,
	ch chan<- prometheus.Metric,
) {
	for _, team := range teams {
		if len(includeTeams) == 0 || existsInSlice(*team.Slug, includeTeams) {
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
	}
}

// exportProjects adds the sentry_project_errors metrics to the collector for export
func exportProjects(
	collector *sentryCollector,
	ch chan<- prometheus.Metric,
) {
	if lastScan["errors"] == 0 {
		lastScan["errors"] = time.Now().Add(time.Second * -10).Unix()
	}
	var wg sync.WaitGroup
	for _, project := range projects {
		queries := []string{"received", "rejected", "blacklisted", "generated"}
		for _, query := range queries {
			wg.Add(1)
			go exportProject(&wg, project, query, collector, ch)
		}
	}
	wg.Wait()
	lastScan["errors"] = time.Now().Unix()
}

// exportProject exports the metrics for a single project and query type
func exportProject(
	wg *sync.WaitGroup,
	p sentry.Project,
	q string,
	collector *sentryCollector,
	ch chan<- prometheus.Metric,
) {
	defer wg.Done()
	if (len(includeProjects) == 0 || existsInSlice(*p.Slug, includeProjects)) &&
		(len(includeTeams) == 0 || isProjectInIncludedTeams(*p.Slug, includeTeams)) {
		count, err := fetchErrorCount(p, q)
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
}

// existsInSlice checks whether a specific string value exists in a string slice
func existsInSlice(value string, slice []string) bool {
	for _, s := range slice {
		if s == value {
			return true
		}
	}
	return false
}

// isProjectInIncludedTeams checks whether at least one of the included teams is linked
// to the project specified
func isProjectInIncludedTeams(projectSlug string, includeTeams []string) bool {
	for _, t := range teams {
		if !existsInSlice(*t.Slug, includeTeams) {
			continue
		}
		for _, p := range *t.Projects {
			if *p.Slug == projectSlug {
				return true
			}
		}
	}
	return false
}

// fetchOrganisation updates the global organisation object from data obtained from the
// Sentry API if the TTL has expired
func fetchOrganisation() {
	now := time.Now().Unix()
	if now-lastScan["organisation"] <= viper.GetInt64("ttl_organisation") {
		return
	}
	log.Info().Msg("Organisation TTL expired, refreshing")
	// Create a Sentry client to query the API
	client, err := GetSentryClient()
	if err != nil {
		log.Error().Err(err).Msg(intialiseSentryError)
	}
	organisation, err = client.GetOrganization(viper.GetString("organisation_name"))
	if err != nil {
		log.Error().
			Err(err).
			Str("organisation", viper.GetString("organisation_name")).
			Msg("Could not fetch organisation")
		apiFailureCallCount++
		return
	}
	lastScan["organisation"] = time.Now().Unix()
	apiSuccessCallCount++
}

// fetchTeams updates the global teams object from data obtained from the
// Sentry API if the TTL has expired
func fetchTeams() {
	now := time.Now().Unix()
	if now-lastScan["teams"] <= viper.GetInt64("ttl_teams") {
		return
	}
	log.Info().Msg("Teams TTL expired, refreshing")
	// Create a Sentry client to query the API
	client, err := GetSentryClient()
	if err != nil {
		log.Error().Err(err).Msg(intialiseSentryError)
	}
	teams, err = client.GetOrganizationTeams(organisation)
	if err != nil {
		log.Error().Err(err).Msg("Could not fetch organisation teams")
		apiFailureCallCount++
		return
	}
	lastScan["teams"] = time.Now().Unix()
	apiSuccessCallCount++
}

// fetchProjects updates the global projects object from data obtained from the
// Sentry API if the TTL has expired
func fetchProjects() {
	now := time.Now().Unix()
	if now-lastScan["projects"] <= viper.GetInt64("ttl_projects") {
		return
	}
	log.Info().Msg("Project TTL expired, refreshing")
	// Create a Sentry client to query the API
	client, err := GetSentryClient()
	if err != nil {
		log.Error().Err(err).Msg(intialiseSentryError)
	}
	projects, _, err = client.GetOrgProjects(organisation)
	if err != nil {
		log.Error().Err(err).Msg("Could not fetch organisation projects")
		apiFailureCallCount++

	}
	apiSuccessCallCount++
	lastScan["projects"] = time.Now().Unix()
}

// fetchErrorCount queries the Sentry API for the error counts of the particular type
// for the specified project.  If there are multiple 10s buckets returned, it adds them
// together to return a single count.
func fetchErrorCount(project sentry.Project, query string) (float64, error) {
	resolution := "10s"
	var err error
	var c []sentry.Stat
	// Create a Sentry client to query the API
	client, err := GetSentryClient()
	if err != nil {
		log.Error().Err(err).Msg(intialiseSentryError)
	}
	// Retry 3 times to fetch stats if there's a failure, with a 3s break between retries
	for i := 0; i < 3; i++ {
		log.Debug().
			Str("project", *project.Slug).
			Str("query", query).
			Int("attempt", i+1).
			Msg("Fetching error counts")
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
			log.Debug().
				Str("project", *project.Slug).
				Str("query", query).
				Msg("Could not fetch stats. Retrying")
			apiFailureCallCount++
			time.Sleep(time.Second * 3)
		} else {
			err = nil
			apiSuccessCallCount++
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
