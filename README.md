# Sentry Exporter

Sentry Exporter exports metrics from your Sentry Organisation and pushes it to a Prometheus push gateway.

## Setup

By default, the application will look for a settings file called `.sentry-exporter.yaml` in the working directory.  An example file is included in this repository.

Alternatively you can specify the location of the settings file with the `--config` command line parameter.

You can also specify any of the settings via ENV variables.

### Settings

| Setting | ENV Var | Default | Description |
| ------- | ------- | ------- | ----------- |
| api_url | SENTRY_EXPORTER_API_URL | nil | URL that points to your Sentry API (typically this is left unspecified and internally defaults to `https://sentry.io/api/0/`)
| organisation_name | SENTRY_EXPORTER_ORGANISATION_NAME | "" | This is a **required** setting. The organisation slug for your account.  This is queried to extract a list of teams
| timeout | SENTRY_EXPORTER_TIMEOUT | nil | The maximum amount of seconds to wait for the Sentry API to respond.  Internally this defaults to 1 minute
| token | SENTRY_EXPORTER_TOKEN | nil | This is a **required** setting. It allows communication with the Sentry API. More details below

### Sentry Token

You need to [create an integration](https://blog.sentry.io/2019/08/21/sentrys-internal-integrations-build-internal-tools-that-fit-your-workflow) to obtain a token which will be used to communicate with the Sentry API.

It will need the following permissions:
* Project - Read
* Team - Read
* Issue & Event - Read
* Organization - Read

## Usage

In order to query the Sentry API and push the metrics to Prometheus, the `sentry-exporter export` command needs to be run.

## What does it do exactly?

The application will query Sentry (using your token) for all the teams in your organisation.  It will then query each team for a list of projects associated with that team.  Finally, it will query each project for a range of stats (received, rejected, blacklisted and generated).  It then pushes these stats to Prometheus with appropriate labels.
