# Sentry Exporter

Sentry Exporter exposes a `/metrics` endpoint that Prometheus can scrape to obtain information about all the projects and teams in your Sentry organisation.

## Setup

By default, the application will look for a settings file called `.sentry-exporter.yaml` in the working directory.  An example file is included in this repository.

Alternatively you can specify the location of the settings file with the `--config` command line parameter.

You can also specify any of the settings via ENV variables.

See the [Configuration](docs/Configuration.md) documentation for further details


## Usage

`sentry-exporter listen` will start a web server that exposes a `/metrics` endpoint that can be scraped by Prometheus.

## Metrics exposed

| Metric | Labels | Detail |
| ------ | ------ | ------ |
| sentry_project_errors | organisation, project, query | Details the number of specific error types (query) encountered by a project in an organisation
| sentry_project_info | organisation, project, team | A purely informational/helper metric to show which teams have which projects associated with them
| sentry_api_calls | status | Details the number of successful or failed calls made to the Sentry API

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
