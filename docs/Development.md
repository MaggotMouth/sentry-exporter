# Development

## Building

`CGO_ENABLED=0 GOOS=linux go build -o bin/sentry-exporter`

## Running the docker image

```
docker build -t sentry-exporter .
docker run --rm --net=<common network name> \
    -p 9142:9142 \
    --name=sentry \
    sentry-exporter listen \
    --loglevel=debug \
    --token=<your sentry token> \
    --organisation=<your sentry organization> \
    --include-queries=generated
```

See the [Configuration](Configuration.md) documentation for a full list of options.

## Additional

You'll also need an instance of Prometheus running to scrape your exporter.

A simple setup is below:

Dockerfile:
```
FROM prom/prometheus

ADD prometheus.yml /etc/prometheus/prometheus.yml
```

prometheus.yml
```
global:
  scrape_interval:     30s
  evaluation_interval: 10s

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
  - job_name: 'sentry'
    static_configs:
      - targets: ['sentry:9142']
```

Finally, here's a little helper script to generate some errors:
```
package main

import (
	"errors"
	"log"
	"math/rand"
	"time"

	"github.com/getsentry/sentry-go"
)

func main() {
	dsns := []string{
		<add your list of project DSNs here>
	}
	rand.Seed(time.Now().Unix())

	for i := true; i; {
		num := rand.Intn(5)
		err := sentry.Init(sentry.ClientOptions{
			Dsn: dsns[num],
			Debug: false,
		})
		if err != nil {
			panic(err)
		}
		randomErr := errors.New("This is a random error")
		log.Printf("Sending error to project %d", num)
		sentry.CaptureException(randomErr)

		sleepyTime := rand.Intn(3)
		time.Sleep(time.Second * time.Duration(sleepyTime))
	}
}

```
