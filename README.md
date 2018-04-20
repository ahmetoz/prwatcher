# prwatcher

- Example usage:

``` shell
prwatcher --host http://stash.example.com --project foo --repository bar--username demo --password 123456 --trigger_uri http://fabrika.example.com:8080/view/Fabrika/job/Demo/buildWithParameters?token=42
```

## help flag.

```shell
prwatcher.exe --help

NAME:
   prwatcher - watch stash pull requests if changes then trigger jenkins - A new cli application

USAGE:
   prwatcher.exe [global options] command [command options] [arguments...]

VERSION:
   0.0.1

DESCRIPTION:
   watch stash pull requests if changes then trigger

AUTHOR:
   Ahmet Oz <bilmuhahmet@gmail.com>

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --host value         host address of docker registry [%HOST%]
   --project value      Only projects for which the authenticated user has the PROJECT_VIEW permission will be returned. [%PROJECT%]
   --repository value   The authenticated user must have REPO_READ permission for the specified project to call this resource. (default: "latest") [%REPOSITORY%]
   --username value     stash user name [%USERNAME%]
   --password value     stash user password [%PASSWORD%]
   --trigger_uri value  job trigger uri - pr id will be added as query string to uri [%TRIGGER_URI%]
   --duration value     job duration https://godoc.org/github.com/robfig/cron#hdr-Intervals (default: "@every 5m") [%DURATION%]
   --help, -h           show help
   --version, -v        print the version

```

## building the code

``` shell
cd prwatcher
go install
```

## to work with docker

```
cd prwatcher
set GOOS=linux
go build
docker-compose up
```
