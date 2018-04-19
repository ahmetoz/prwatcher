# prwatcher

- Example usage:

``` shell
prwatcher --host http://stash.example.com --project foo --repository bar--username demo --password 123456 --trigger_uri http://fabrika.example.com:8080/view/Fabrika/job/Demo/buildWithParameters?token=42
```

```shell

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
   --host value         host address of docker registry
   --project value      Only projects for which the authenticated user has the PROJECT_VIEW permission will be returned.
   --repository value   The authenticated user must have REPO_READ permission for the specified project to call this resource. (default: "latest")
   --username value     registry user name
   --password value     registry user password
   --trigger_uri value  job trigger uri - pr id will be added as query string to uri
   --pre_trigger value  check uri before triggering job
   --duration value     job duration https://godoc.org/github.com/robfig/cron#hdr-Intervals (default: "@every 5s")
   --help, -h           show help
   --version, -v        print the version

```

## building the code

``` shell
cd prwatcher
go install
```
