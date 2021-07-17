# Sentry Exporter

Sentry Exporter exposes a `/metrics` endpoint that Prometheus can scrape to obtain information about all the projects and teams in your Sentry organisation.

## Setup

By default, the application will look for a settings file called `.sentry-exporter.yaml` in the working directory.  An example file is included in this repository.

Alternatively you can specify the location of the settings file with the `--config` command line parameter.

You can also specify any of the settings via ENV variables.

### Settings file

| Setting | ENV Var | Default | Description |
| ------- | ------- | ------- | ----------- |
| api_url | SENTRY_EXPORTER_API_URL | https://sentry.io/api/0/ | URL that points to your Sentry API
| listen_address | SENTRY_EXPORTER_LISTEN_ADDRESS | :9142 | Address to start the web server on
| organisation_name | SENTRY_EXPORTER_ORGANISATION_NAME | "" | This is a **required** setting. The organisation slug for your account.  This is queried to extract a list of teams
| project_includes | SENTRY_EXPORTER_PROJECT_INCLUDES | "" | Comma separated list of project slugs to include in the export
| team_includes | SENTRY_EXPORTER_TEAM_INCLUDES | "" | Comma separated list of team slugs to include in the export
| timeout | SENTRY_EXPORTER_TIMEOUT | 60 | The maximum amount of seconds to wait for the Sentry API to respond
| token | SENTRY_EXPORTER_TOKEN | "" | This is a **required** setting. It allows communication with the Sentry API. More details below
| ttl_organisation | SENTRY_EXPORTER_TTL_ORGANISATION | 86400 | The duration in seconds to hold organisation information in memory (no request to Sentry)
| ttl_projects | SENTRY_EXPORTER_TTL_PROJECTS | 600 | The duration in seconds to hold project information in memory (no request to Sentry)
| ttl_teams | SENTRY_EXPORTER_TTL_TEAMS | 3600 | The duration in seconds to hold team information in memory (no request to Sentry)


### Command line parameters

| Flag | Default | Description |
| ---- | ------- | ----------- |
| --config | $CURRENT_DIR/.sentry-exporter.yaml | Location of the configuration file to load
| --loglevel | info | What level of logs should be exposed.  Options are trace, debug, info, warn, error, fatal or panic
| --logformat | text | What format should logs be output as, human-friendly text, or computer-friendly json. Options are text or json
| --token | "" | Allows you to specify the Sentry token via parameter instead of in config file


### Sentry Token

You need to [create an integration](https://blog.sentry.io/2019/08/21/sentrys-internal-integrations-build-internal-tools-that-fit-your-workflow) to obtain a token which will be used to communicate with the Sentry API.

It will need the following permissions:
* Project - Read
* Team - Read
* Issue & Event - Read
* Organization - Read

## Usage

`sentry-exporter listen` will start a web server.

## Metrics exposed

| Metric | Labels | Detail |
| ------ | ------ | ------ |
| sentry_project_errors | organisation, project, query | Details the number of specific error types (query) encountered by a project in an organisation
| sentry_project_info | organisation, project, team | A purely informational/helper metric to show which teams have which projects associated with them

Example PromQL query showing the number of errors received for a particular team, broken down by project:
```
sentry_project_errors{query="received"} * on (project) group_left(team) sentry_project_info{team="example-team-1"}
```


## Thanks

Thanks to the contributors of the following projects, without whom this project would not be possible:
* [Cobra & Viper](https://github.com/spf13/cobra)
* [Atlassian Sentry API](https://github.com/atlassian/go-sentry-api)
* [Zerolog](https://github.com/rs/zerolog)


[![Sentry](assets//sentry-wordmark-dark-202x60.png "Sentry")](https://sentry.io/)

Special thanks to [Sentry.io](https://sentry.io/) for providing me with a sponsored account that I can thrash with errors for testing purposes.
