package sentry

import (
	"fmt"
	"time"

	atlassianSentry "github.com/atlassian/go-sentry-api"
	"github.com/rs/zerolog/log"
)

// SentryClientInterface defines the interface that wraps the Atlassian Sentry Go package
type SentryClientInterface interface {
	GetToken() string
	GetOrganisation(string) atlassianSentry.Organization
	GetOrganisationTeams(orgObj atlassianSentry.Organization) []atlassianSentry.Team
	GetTeamProjects(orgObj atlassianSentry.Organization, teamObj atlassianSentry.Team) []atlassianSentry.Project
	GetProjectStats(orgObj atlassianSentry.Organization,
		prjObj atlassianSentry.Project,
		query string,
	) []atlassianSentry.Stat
}

// SentryClient implements SentryClientInterface
type SentryClient struct {
	token  string
	client *atlassianSentry.Client
}

// NewSentryClient instantiates a client to communicate with the Sentry API
func NewSentryClient(token string, url *string, timeout *int) SentryClientInterface {
	log.Debug().Msg("Instantiating Sentry Client")
	client, err := atlassianSentry.NewClient(
		token,
		url,
		timeout,
	)
	if err != nil {
		log.Error().Err(err).Msg("Could not initialise Sentry Client")
	}
	return SentryClient{
		token:  token,
		client: client,
	}
}

// GetToken returns the token used to access the Sentry API
func (s SentryClient) GetToken() string {
	return s.token
}

// GetOrganisation queries the Sentry API to fetch the Organisation object
// with the name specified
func (s SentryClient) GetOrganisation(org string) atlassianSentry.Organization {
	log.Debug().
		Str("org", org).
		Msg("Querying API for Organisation")
	orgObj, err := s.client.GetOrganization(org)
	if err != nil {
		log.Error().
			Err(err).
			Str("org", org).
			Msg("Could not fetch organisation")
	}
	return orgObj
}

// GetOrganisationTeams queries the Sentry API to fetch the Team objects
// linked to the organisation
func (s SentryClient) GetOrganisationTeams(orgObj atlassianSentry.Organization) []atlassianSentry.Team {
	log.Debug().
		Str("org", *orgObj.Slug).
		Msg("Querying API for Organisation Teams")
	teams, err := s.client.GetOrganizationTeams(orgObj)
	if err != nil {
		log.Error().
			Err(err).
			Str("org", orgObj.Name).
			Msg("Could not fetch teams for organisation")
	}
	return teams
}

// GetTeamProjects queries the Sentry API to fetch all the Project objects
// that are linked to the team
func (s SentryClient) GetTeamProjects(
	orgObj atlassianSentry.Organization,
	teamObj atlassianSentry.Team,
) []atlassianSentry.Project {
	log.Debug().
		Str("org", *orgObj.Slug).
		Str("team", *teamObj.Slug).
		Msg("Querying API for Organisation Team Projects")
	projects, err := s.client.GetTeamProjects(orgObj, teamObj)
	if err != nil {
		log.Error().
			Err(err).
			Str("org", orgObj.Name).
			Str("team", teamObj.Name).
			Msg("Could not fetch projects for team")
	}
	return projects
}

// GetProjectStats queries the Sentry API to fetch all the stats for the Project
func (s SentryClient) GetProjectStats(
	orgObj atlassianSentry.Organization,
	prjObj atlassianSentry.Project,
	query string,
) []atlassianSentry.Stat {
	log.Debug().
		Str("org", *orgObj.Slug).
		Str("project", *prjObj.Slug).
		Msg("Querying API for Organisation Team Project Stats")

	from := time.Now().Add(time.Minute * -5)
	resolution := "10s"

	received, err := s.client.GetProjectStats(
		orgObj,
		prjObj,
		atlassianSentry.StatQuery(query),
		from.Unix(),
		time.Now().Unix(),
		&resolution,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("org", orgObj.Name).
			Str("project", prjObj.Name).
			Str("query", query).
			Msg("Could not fetch stats for project")
	}
	return received
}

type SentryStat struct {
	Query  string
	Counts []atlassianSentry.Stat
}

type SentryProject struct {
	Project atlassianSentry.Project
	Stats   []SentryStat
}

type SentryTeam struct {
	Team     atlassianSentry.Team
	Projects []SentryProject
}

type SentryOrganisation struct {
	Organisation atlassianSentry.Organization
	Teams        []SentryTeam
}

func (s SentryOrganisation) String() string {
	var str string
	for _, team := range s.Teams {
		for _, project := range team.Projects {
			for _, stat := range project.Stats {
				str += fmt.Sprintf("Organisation: %s - ", *s.Organisation.Slug)
				str += fmt.Sprintf("Team: %s - ", *team.Team.Slug)
				str += fmt.Sprintf("Project: %s - ", *project.Project.Slug)
				str += fmt.Sprintf("Query: %s - ", stat.Query)
				str += fmt.Sprintf("Counts: Slice of length(%d)", len(stat.Counts))
				str += " --- "
			}
		}
	}
	return str
}

// SentryProcessor refines the output from a Sentry Client into the
// data we want for the exporter
type SentryProcessor struct {
	Client SentryClientInterface
}

// NewSentryProcessor returns a SentryProcessor object
func NewSentryProcessor(client SentryClientInterface) SentryProcessor {
	return SentryProcessor{
		Client: client,
	}
}

// GetStats gets all the stats
func (sp SentryProcessor) GetStats(org string) SentryOrganisation {
	orgObj := sp.Client.GetOrganisation(org)
	teams := sp.Client.GetOrganisationTeams(orgObj)
	queries := []string{"received", "rejected", "blacklisted", "generated"}
	result := SentryOrganisation{
		Organisation: orgObj,
	}
	for _, team := range teams {
		t := SentryTeam{
			Team: team,
		}
		for _, project := range sp.Client.GetTeamProjects(orgObj, team) {
			p := SentryProject{
				Project: project,
			}
			for _, query := range queries {
				stats := sp.Client.GetProjectStats(orgObj, project, query)
				q := SentryStat{
					Query:  query,
					Counts: stats,
				}
				p.Stats = append(p.Stats, q)
			}
			t.Projects = append(t.Projects, p)
		}
		result.Teams = append(result.Teams, t)
	}
	return result
}
