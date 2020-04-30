# Camunda-CI Dashboard

This package provides a biased Dashboard for Camunda-CIs Jenkins & Travis Broken Jobs board

## Build

Requirements:
* Go 1.8
* Make
* Docker

```
git clone https://github.com/camunda-ci/camunda-ci-dashboard
cd camuda-ci-dashboard
make package
```

## Usage

Docker Image:
```
docker run -t -e CCD_USERNAME=foo -e CCD_PASSWORD=bar registry.camunda.com/camunda-ci-dashboard:latest
```

Binary:
```
./camunda-ci-dashboard [--debug=true] --username=foo --password=bar --bindAddress=0.0.0.0:8000
```

## Configuration

The Jenkins Username and Jenkins Password for Basic Auth can be set either using the cmdline flags, inside the `.camunda-ci-dashboard.json` config file or specified as environment variables.

* `CCD_USERNAME`
* `CCD_PASSWORD`
* `CCD_BINDADDRESS`
* `CCD_DEBUG`

## Example Config

```json
{
	"username": "<jenkins username>",
	"password": "<jenkins password>",
	"jenkins": {
		"Release": {
			"url": "http://release-cambpm-ui:8080",
			"publicUrl": "https://release.cambpm.camunda.cloud"
		},
		"Docs": {
			"url": "http://ci-cambpm-ui:8080",
			"brokenJobsUrl": "http://ci-cambpm-ui:8080/job/docs",
			"publicUrl": "https://ci.cambpm.camunda.cloud/job/docs"
		}
	},
	"travis": {
		"accessToken": "",
		"organizations": [
			{
				"name": "camunda",
				"repos": [
					{
						"name": "camunda-external-task-client-js"
					},
					{
						"name": "camunda-bpm-assert-scenario"
					}
				]
			}
		]
	}
}
```
