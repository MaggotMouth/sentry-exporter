# Configuration

## Settings file

| Setting | ENV Var | Default | Description |
| ------- | ------- | ------- | ----------- |
| api_url | SENTRY_EXPORTER_API_URL | https://sentry.io/api/0/ | URL that points to your Sentry API
| listen_address | SENTRY_EXPORTER_LISTEN_ADDRESS | :9142 | Address to start the web server on
| organisation_name | SENTRY_EXPORTER_ORGANISATION_NAME | "" | This is a **required** setting. The organisation slug for your account.  This is queried to extract a list of teams
| include_projects | SENTRY_EXPORTER_INCLUDE_PROJECTS | "" | Comma separated list of project slugs to include in the export
| include_queries | SENTRY_EXPORTER_INCLUDE_QUERIES | "" | Comma separated list of query types to include in the export (valid options: generated, blacklisted, received, rejected)
| include_teams | SENTRY_EXPORTER_INCLUDE_TEAMS | "" | Comma separated list of team slugs to include in the export
| timeout | SENTRY_EXPORTER_TIMEOUT | 60 | The maximum amount of seconds to wait for the Sentry API to respond
| token | SENTRY_EXPORTER_TOKEN | "" | This is a **required** setting. It allows communication with the Sentry API. More details below
| ttl_organisation | SENTRY_EXPORTER_TTL_ORGANISATION | 86400 | The duration in seconds to hold organisation information in memory (no request to Sentry)
| ttl_projects | SENTRY_EXPORTER_TTL_PROJECTS | 600 | The duration in seconds to hold project information in memory (no request to Sentry)
| ttl_teams | SENTRY_EXPORTER_TTL_TEAMS | 3600 | The duration in seconds to hold team information in memory (no request to Sentry)


## Command line parameters

| Flag | Default | Description |
| ---- | ------- | ----------- |
| --config | $CURRENT_DIR/.sentry-exporter.yaml | Location of the configuration file to load
| --include-projects | "" | Which projects should be included in the export
| --include-queries | "" | Which query types should be included in the export.  Options are generated, blacklisted, received, or rejected
| --include-teams | "" | Which teams should be included in the export
| --loglevel | info | What level of logs should be exposed.  Options are trace, debug, info, warn, error, fatal or panic
| --logformat | text | What format should logs be output as, human-friendly text, or computer-friendly json. Options are text or json
| --token | "" | Allows you to specify the Sentry token via parameter instead of in config file


## Sentry Token

You need to [create an integration](https://blog.sentry.io/2019/08/21/sentrys-internal-integrations-build-internal-tools-that-fit-your-workflow) to obtain a token which will be used to communicate with the Sentry API.

It will need the following permissions:
* Project - Read
* Team - Read
* Issue & Event - Read
* Organization - Read
