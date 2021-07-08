package sentry

import (
	"math/rand"
	"testing"
	"time"

	atlassianSentry "github.com/atlassian/go-sentry-api"
)

// TestSentryClient is a mock for the Sentry Client
// Implements SentryInterface
type TestSentryClient struct {
	token string
}

// GetToken returns the token used to access the Sentry API
func (s TestSentryClient) GetToken() string {
	return s.token
}

// GetOrganisation queries the Sentry API to fetch the Organisation object
// with the name specified
func (s TestSentryClient) GetOrganisation(org string) atlassianSentry.Organization {
	n := "mock-organisation"
	return atlassianSentry.Organization{
		Name: org,
		Slug: &n,
	}
}

// GetOrganisationTeams queries the Sentry API to fetch the Team objects
// linked to the organisation
func (s TestSentryClient) GetOrganisationTeams(orgObj atlassianSentry.Organization) []atlassianSentry.Team {
	n1 := "mock-team-1"
	n2 := "mock-team-2"
	n3 := "mock-team-3"
	return []atlassianSentry.Team{
		{
			Name: "Mock Team 1",
			Slug: &n1,
		},
		{
			Name: "Mock Team 2",
			Slug: &n2,
		},
		{
			Name: "Mock Team 3",
			Slug: &n3,
		},
	}
}

// GetTeamProjects queries the Sentry API to fetch all the Project objects
// that are linked to the team
func (s TestSentryClient) GetTeamProjects(
	orgObj atlassianSentry.Organization,
	teamObj atlassianSentry.Team,
) []atlassianSentry.Project {
	n1 := "mock-project-1"
	n2 := "mock-project-2"
	n3 := "mock-project-3"
	return []atlassianSentry.Project{
		{
			Name: "Mock Project 1",
			Slug: &n1,
		},
		{
			Name: "Mock Project 2",
			Slug: &n2,
		},
		{
			Name: "Mock Project 3",
			Slug: &n3,
		},
	}
}

// GetProjectStats queries the Sentry API to fetch all the stats for the Project
func (s TestSentryClient) GetProjectStats(
	orgObj atlassianSentry.Organization,
	prjObj atlassianSentry.Project,
	query string,
) []atlassianSentry.Stat {
	timestamp := time.Now().Add(-60 * time.Second)
	var result []atlassianSentry.Stat
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 6; i++ {
		num := rand.Int63n(1000)
		result = append(result, atlassianSentry.Stat{float64(num), float64(timestamp.Unix())})
		timestamp = timestamp.Add(10 * time.Second)
	}
	return result
}

var processor SentryProcessor

const testOrgName = "test"

func newSentryProcessor() {
	if processor.Client == nil {
		client := TestSentryClient{token: "ABCD-EFGH-IJKL-MNOP"}
		processor = NewSentryProcessor(client)
	}
}

func TestGetStats(t *testing.T) {
	newSentryProcessor()
	stats := processor.GetStats(testOrgName)
	if stats.Organisation.Name != testOrgName {
		t.Errorf("Expected organisation name = %s, got %s", testOrgName, stats.Organisation.Name)
	}
}

func TestString(t *testing.T) {
	newSentryProcessor()
	stats := processor.GetStats(testOrgName)
	want := "Organisation: mock-organisation - Team: mock-team-1 - Project: mock-project-1 - Query: received - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-1 - Project: mock-project-1 - Query: rejected - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-1 - Project: mock-project-1 - Query: blacklisted - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-1 - Project: mock-project-1 - Query: generated - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-1 - Project: mock-project-2 - Query: received - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-1 - Project: mock-project-2 - Query: rejected - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-1 - Project: mock-project-2 - Query: blacklisted - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-1 - Project: mock-project-2 - Query: generated - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-1 - Project: mock-project-3 - Query: received - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-1 - Project: mock-project-3 - Query: rejected - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-1 - Project: mock-project-3 - Query: blacklisted - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-1 - Project: mock-project-3 - Query: generated - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-2 - Project: mock-project-1 - Query: received - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-2 - Project: mock-project-1 - Query: rejected - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-2 - Project: mock-project-1 - Query: blacklisted - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-2 - Project: mock-project-1 - Query: generated - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-2 - Project: mock-project-2 - Query: received - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-2 - Project: mock-project-2 - Query: rejected - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-2 - Project: mock-project-2 - Query: blacklisted - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-2 - Project: mock-project-2 - Query: generated - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-2 - Project: mock-project-3 - Query: received - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-2 - Project: mock-project-3 - Query: rejected - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-2 - Project: mock-project-3 - Query: blacklisted - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-2 - Project: mock-project-3 - Query: generated - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-3 - Project: mock-project-1 - Query: received - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-3 - Project: mock-project-1 - Query: rejected - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-3 - Project: mock-project-1 - Query: blacklisted - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-3 - Project: mock-project-1 - Query: generated - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-3 - Project: mock-project-2 - Query: received - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-3 - Project: mock-project-2 - Query: rejected - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-3 - Project: mock-project-2 - Query: blacklisted - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-3 - Project: mock-project-2 - Query: generated - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-3 - Project: mock-project-3 - Query: received - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-3 - Project: mock-project-3 - Query: rejected - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-3 - Project: mock-project-3 - Query: blacklisted - Counts: Slice of length(6) --- Organisation: mock-organisation - Team: mock-team-3 - Project: mock-project-3 - Query: generated - Counts: Slice of length(6) --- "
	if stats.String() != want {
		t.Errorf("Wrong output for String()")
	}
}
