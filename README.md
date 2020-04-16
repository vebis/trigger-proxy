# trigger-proxy

trigger-proxy is used to map git repository / branch tuples to job names on Jenkins. Especially, but not exclusive for multibranch pipeline projects.
This is useful to have commit hook triggered builds on pipelines without direct git association in Jenkins.

## Badges

[![Go Report Card](https://goreportcard.com/badge/github.com/vebis/trigger-proxy)](https://goreportcard.com/report/github.com/vebis/trigger-proxy)

## Getting started

Set the following command line parameters

* jenkins-url - your jenkins installation
* jenkins-multi - name of multibranch pipeline project
* jenkins-user - user who can trigger builds
* jenkins-token - the api token of the user
* quietperiod - quiet period for jobs, defaults to 30 (seconds)
* mapping-file - path to mapping file, defaults to mapping.csv
* mapping-url - path to mapping file on an http server (sha256 hash of mapping file at same url with .sha256 suffix)
* mappingrefresh - intervall to check for changed mappings, defaults to 5 (minutes)
* filematch - parses a 4th column of the mapping file and tries to match files received in the request
* semanticrepo - semantic repos, a corner case, you know if you need this (component/package setups). If this parameter is defined, filematch is set to true!
* port -  http port to listen on (defaults to 8080)

## Usage

```bash
sudo docker run vebis/trigger-proxy \
    -jenkins-url="https://jenkins:8443" \
    -jenkins-user="triggeruser" \
    -jenkins-token="token"
```

Send an http GET request with parameter "repo" to port 8080. If you defined the parameter "branch" it will be considered, otherwise "master" ist assumed.
If you send one or more paramters "file" and you have filematching enabled, it tries to match the files provided against your mapping.
Otherwise, if you send a [GitLab Webhook](https://docs.gitlab.com/ee/user/project/integrations/webhooks.html) to the endpoint "/json", the information will be parsed and matched against your mapping.
The app will lookup any job names for your input and will trigger them.

### Use Case - monorepo

If you have a monorepo and want to trigger specific builds, you can do this easily.

Example mappingfile:

```csv
https://gitserver/monorepo.git,master,jenkinsjobproj1,subdir1
https://gitserver/monorepo.git,master,jenkinsjobproj2,subdir2
```

```bash
sudo docker run vebis/trigger-proxy \
    -jenkins-url="https://jenkins:8443" \
    -jenkins-user="triggeruser" \
    -jenkins-token="token"
    -filematch
```

If you send an HTTP GET request like:

```bash
curl http://trigger-proxy:8080/?repo=https://gitserver/monorepo.git\&branch=master\&file=subdir2/README.md
```

Jenkins job "jenkinsjobproj2" will be triggered.

### Use Case - semantic repo

Sometime it happends you have a special meaning in the path component of your git repo. Like when you have a component which consists of multiple packages.

Example mappingfile:

```csv
https://gitserver/components/component1.git,master,jenkinsjobproj1,component1/package1
https://gitserver/components/component1.git,master,jenkinsjobproj2,component1/package2
https://gitserver/components/component2.git,master,jenkinsjobproj2,component1/package1
https://gitserver/monorepo.git,master,jenkinsjobproj2,subdir2
```

```bash
sudo docker run vebis/trigger-proxy \
    -jenkins-url="https://jenkins:8443" \
    -jenkins-user="triggeruser" \
    -jenkins-token="token"
    -filematch
    -semanticrepo="https://gitserver/components/"
```

If you send an HTTP GET request like:

```bash
curl http://trigger-proxy:8080/?repo=https://gitserver/components/component1.git\&branch=master\&file=package1/README.md
```

Then just your Jenkins job "jenkinsjobproj1" will be triggered.

## Misc

There is a readiness endpoint at "/readyz".

## Authors

* **Stephan Kirsten**

## License

BSD 2-Clause "Simplified" License
