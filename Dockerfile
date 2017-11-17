FROM golang:1.11-alpine as builder
RUN apk add --no-cache bash build-base curl findutils git make tar xz
WORKDIR ${GOPATH}/src/github.com/camunda-ci/camunda-ci-dashboard
COPY . ${GOPATH}/src/github.com/camunda-ci/camunda-ci-dashboard
RUN make build distribution

FROM golang:1.11-alpine
COPY --from=builder ${GOPATH}/src/github.com/camunda-ci/camunda-ci-dashboard/bin/camunda-ci-dashboard_linux_amd64 /camunda-ci-dashboard
RUN chmod +x /camunda-ci-dashboard && \
    addgroup -S app && \
    adduser -S -g app app && \
        chown app:app /camunda-ci-dashboard
ENTRYPOINT ["/camunda-ci-dashboard"]
CMD ["--bindAddress", "0.0.0.0:8000"]
