FROM golang:alpine

COPY ./bin/sentry-exporter /

ENTRYPOINT ["/sentry-exporter"]
CMD ["listen", "--loglevel=debug"]
