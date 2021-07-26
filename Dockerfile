FROM golang:alpine

COPY ./bin/sentry-exporter /
RUN chmod a+x /sentry-exporter

ENTRYPOINT ["/sentry-exporter"]
CMD ["listen", "--loglevel=debug"]
