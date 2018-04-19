package main

import (
	"encoding/base64"
	"encoding/json"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"fmt"
)

type PullRequestsResponse struct {
	Size   int `json:"size"`
	Values []struct {
		ID          int    `json:"id"`
		Version     int    `json:"version"`
		State       string `json:"state"`
		Open        bool   `json:"open"`
		Closed      bool   `json:"closed"`
		CreatedDate int64  `json:"createdDate"`
		UpdatedDate int64  `json:"updatedDate"`
	} `json:"values"`
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func getPullRequestsResponse(body []byte) (*PullRequestsResponse, error) {
	s := new(PullRequestsResponse)
	err := json.Unmarshal(body, &s)
	if err != nil {
		log.Error(err)
	}
	return s, err
}

func getPullRequests(host, project, repository, token string) *PullRequestsResponse {
	url := host + "/rest/api/1.0/projects/" + project + "/repos/" + repository + "/pull-requests/"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Basic "+token)
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if res.StatusCode != 200 {
		log.Error("Getting pull requests, failed. Status: ", res.StatusCode)
	}

	if res.StatusCode == 403 || res.StatusCode == 401 {
		panic(string(body))
	}

	fmt.Println(string(body))

	if err != nil {
		log.Error(err)
	}

	s, _ := getPullRequestsResponse([]byte(body))
	return s
}

var active_pr_list map[int]int64 = make(map[int]int64)
var passive_pr_list map[int]int64 = make(map[int]int64)

var triggerCount int = 0

func triggerJob(host, project, repository, token, trigger_uri string) {
	pull_requests := getPullRequests(host, project, repository, token)
	triggerList := getTriggerList(pull_requests)
	for _, id := range triggerList {
		log.Info("PR: ", id, " trigger starting")
		triggerToUri(id, trigger_uri)
	}
}

func getTriggerList(pull_requests *PullRequestsResponse) []int {
	triggerCount++
	if triggerCount%10 == 0 {
		log.Info("triggered ", triggerCount, " times")
		log.Info("active pull requests: ", active_pr_list)
		log.Info("passive pull requests: ", passive_pr_list)
	}
	var triggerList []int
	for k, v := range active_pr_list {
		if v == 0 {
			continue
		}

		found := false
		for _, pr := range pull_requests.Values {
			if k == pr.ID {
				found = true
				break
			}
		}

		if !found {
			delete(active_pr_list, k)
			passive_pr_list[k] = v
			log.Info("PR: ", k, " is close(aka passive).")
		}
	}

	for _, pr := range pull_requests.Values {

		updateDate, exists := active_pr_list[pr.ID]
		if exists {
			if updateDate == 0 {
				log.Info("PR: ", pr.ID, " will be tried again.")
				active_pr_list[pr.ID] = pr.UpdatedDate
			}
			if pr.UpdatedDate == updateDate {
				log.Info("PR: ", pr.ID, " open but this version triggered before. Will be skipped.")
			} else {
				log.Info("PR: ", pr.ID, " triggered before but updated on  ", pr.UpdatedDate, " so will be triggered again.")
				active_pr_list[pr.ID] = pr.UpdatedDate
				triggerList = append(triggerList, pr.ID)
			}
		} else {
			log.Info("New PR: ", pr.ID, " created on ", pr.CreatedDate, " updated on ", pr.UpdatedDate)
			active_pr_list[pr.ID] = pr.UpdatedDate
			triggerList = append(triggerList, pr.ID)
		}
	}

	return triggerList
}

func triggerToUri(id int, trigger_uri string) {
	url := trigger_uri + "&cause=pr-watcher&pr=" + strconv.Itoa(id)
	log.Info("sending post request to ", url)

	req, _ := http.NewRequest("POST", url, nil)

	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != 201 {
		log.Error("PR: " , id, " trigger failed. Status: ", res.StatusCode)
		active_pr_list[id] = 0 //0 means try later
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	if res.StatusCode == 403 || res.StatusCode == 401 {
		panic(string(body))
	}

	fmt.Println(string(body))
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	app := cli.NewApp()
	app.Name = "prwatcher - watch stash pull requests if changes then trigger jenkins"
	app.Description = "watch stash pull requests if changes then trigger"
	app.Version = "0.0.1"
	app.Authors = []cli.Author{
		{
			Name:  "Ahmet Oz",
			Email: "bilmuhahmet@gmail.com",
		},
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "host",
			Value: "",
			Usage: "host address of docker registry",
			EnvVar: "HOST",
		},
		cli.StringFlag{
			Name:  "project",
			Value: "",
			Usage: "Only projects for which the authenticated user has the PROJECT_VIEW permission will be returned.",
			EnvVar: "PROJECT",
		},
		cli.StringFlag{
			Name:  "repository",
			Value: "latest",
			Usage: "The authenticated user must have REPO_READ permission for the specified project to call this resource.",
			EnvVar: "REPOSITORY",
		},
		cli.StringFlag{
			Name:  "username",
			Value: "",
			Usage: "stash user name",
			EnvVar: "USERNAME",
		},
		cli.StringFlag{
			Name:  "password",
			Value: "",
			Usage: "stash user password",
			EnvVar: "PASSWORD",
		},
		cli.StringFlag{
			Name:  "trigger_uri",
			Value: "",
			Usage: "job trigger uri - pr id will be added as query string to uri",
			EnvVar: "TRIGGER_URI",
		},
		cli.StringFlag{
			Name:  "duration",
			Value: "@every 5m",
			Usage: "job duration https://godoc.org/github.com/robfig/cron#hdr-Intervals",
			EnvVar: "DURATION",
		},
	}

	app.Action = func(c *cli.Context) error {
		if c.String("host") == "" {
			return cli.NewExitError("host is required", 86)
		}
		if c.String("project") == "" {
			return cli.NewExitError("project is required", 86)
		}
		if c.String("repository") == "" {
			return cli.NewExitError("project is required", 86)
		}
		if c.String("username") == "" {
			return cli.NewExitError("user name is required", 86)
		}
		if c.String("password") == "" {
			return cli.NewExitError("password is required", 86)
		}
		if c.String("trigger_uri") == "" {
			return cli.NewExitError("trigger_uri is required", 86)
		}

		token := basicAuth(c.String("username"), c.String("password"))
		host := c.String("host")
		project := c.String("project")
		repository := c.String("repository")
		duration := c.String("duration")
		trigger_uri := c.String("trigger_uri")

		log.WithFields(log.Fields{
			"host":        host,
			"project":     project,
			"repository":  repository,
			"duration":    duration,
			"trigger_uri": trigger_uri,
		}).Info("The PR watcher starting...")

		triggerJob(host, project, repository, token, trigger_uri)

		job := cron.New()
		job.AddFunc(duration, func() {
			triggerJob(host, project, repository, token, trigger_uri)
		})
		job.Start()

		for true {

		}

		return nil
	}
	app.Run(os.Args)

}
