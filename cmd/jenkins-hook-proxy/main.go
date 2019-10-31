package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/nlopes/slack"
	"gopkg.in/rjz/githubhook.v0"
	"gopkg.in/src-d/go-errors.v1"
	"gopkg.in/src-d/go-log.v1"
)

const defaultEndpoint = "https://jenkins.k8s.pipeline.prod.srcd.run/generic-webhook-trigger/invoke"

// event is a struct that contains required info from github push event: ref and repository metadata
type event struct {
	Ref  string `json:"ref"`
	Repo repo   `json:"repository"`
}

// repo contains repository name
type repo struct {
	Name string `json:"name"`
}

// helper contains required credentials to parse the request and notify about events
type helper struct {
	slackClient     *slack.Client
	slackChannel    string
	jenkinsEndpoint string
	githubSecret    string
}

type handler func(res http.ResponseWriter, req *http.Request)

// newHelper is a helper constructor
func newHelper() *helper {
	return &helper{
		jenkinsEndpoint: getEnv("JENKINS_ENDPOINT", defaultEndpoint),
		slackChannel:    os.Getenv("SLACK_CHANNEL"),
		slackClient:     slack.New(os.Getenv("SLACK_TOKEN")),
		githubSecret:    os.Getenv("GITHUB_SECRET"),
	}
}

func main() {
	log.Infof("initializing helper")
	helper := newHelper()

	log.Infof("start listening")
	http.HandleFunc("/", helper.handleHookPushEvent())
	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Errorf(err, "server error")
	}
}

// handleHookPushEvent does several steps:
// 1) tries to parse received event from github
// 2) adds token URL variable
// 3) sends request to the Jenkins endpoint
func (h *helper) handleHookPushEvent() handler {
	return func(res http.ResponseWriter, req *http.Request) {
		hook, err := githubhook.Parse([]byte(h.githubSecret), req)
		if err != nil {
			h.logError(err, "failed to parse the request, check your secret")
			return
		}

		data, err := prettyPrint(hook.Payload)
		if err != nil {
			h.logError(err, "failed to prettify data")
			return
		}
		log.Debugf(string(data))

		e := new(event)
		if err := json.Unmarshal(data, e); err != nil {
			h.logError(err, "failed to trigger jenkins")
			return
		}

		if err := h.triggerJenkins(e, data); err != nil {
			h.logError(err, fmt.Sprintf("`%v` `%v` failed to trigger jenkins", e.Repo.Name, e.Ref))
			return
		}
	}
}

// triggerJenkins sends request to the Jenkins endpoint
func (h *helper) triggerJenkins(e *event, data []byte) error {
	URL, err := url.Parse(h.jenkinsEndpoint)
	if err != nil {
		return err
	}
	parameters := url.Values{}
	parameters.Add("token", e.Repo.Name)
	URL.RawQuery = parameters.Encode()
	log.Debugf("URL: %s", URL.String())

	resp, err := http.Post(URL.String(), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.NewKind("something went wrong").Wrap(err)
	}

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	h.logEvent(e.Repo.Name, e.Ref, "%v\njenkins has been successfully triggered", string(respData))
	return nil
}

// logEvent logs corresponding event to Stdout and a slack channel(if configured)
func (h *helper) logEvent(repo, ref, format string, args ...interface{}) {
	log.Infof(format, args...)

	text := fmt.Sprintf(format, args...)
	if h.slackChannel != "" {
		text = fmt.Sprintf(":ok: `%v` `%v` message:\n```%v```", repo, ref, text)
		if _, _, sendErr := h.slackClient.PostMessage(h.slackChannel, slack.MsgOptionText(text, false)); sendErr != nil {
			log.Errorf(sendErr, "failed to send to slack")
		}
	}
}

// logError logs corresponding error to Stdout and a slack channel(if configured)
func (h *helper) logError(err error, format string) {
	log.Errorf(err, format)
	if h.slackChannel != "" {
		text := fmt.Sprintf(":red_circle: %s:\n```\n%v\n```", format, err)
		if _, _, sendErr := h.slackClient.PostMessage(h.slackChannel, slack.MsgOptionText(text, false)); sendErr != nil {
			log.Errorf(sendErr, "failed to send to slack")
		}
	}
}

func prettyPrint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
