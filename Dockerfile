FROM golang:alpine

COPY bin/camunda-ci-dashboard_linux_amd64 /bin/camunda-ci-dashboard

ENTRYPOINT ["/bin/camunda-ci-dashboard"]

CMD ["--bindAddress", "0.0.0.0:8000"]
