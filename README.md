# Camunda-CI Dashboard

This package provides a biased Dashboard for Camunda-CIs Jenkins Broken Jobs board

## Build

Requirements:
* Go 1.8
* Make
* Docker

```
git clone https://github.com/camunda-ci/camunda-ci-dashboard
cd camuda-ci-dashboard
make distribution
```

## Usage

```
./camunda-ci-dashboard [--debug=true] --username=foo --password=bar --bindAddress=0.0.0.0:8000
```

## Configuration

The Jenkins Username and Jenkins Password for Basic Auth can be set either using the cmdline flags, inside the `.camunda-ci-dashboard.json` config file or specified as environment variables.

* CCD_USERNAME
* CCD_PASSWORD
* CCD_BINDADDRESS
* CCD_DEBUG
