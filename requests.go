package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
)

func parseGetRequest(r *http.Request, filematch bool) (string, string, []string, error) {
	repo := ""
	branch := ""
	files := []string{}

	log.Print("parsing get request")
	reqRepo, ok := r.URL.Query()["repo"]

	if !ok || len(reqRepo) < 1 {
		log.Print("repo is missing")
		log.Print("aborting request handling")

		return repo, branch, files, errors.New("repo is missing")
	}

	repo = reqRepo[0]

	log.Print("parsed repo: ", repo)

	reqBranch, ok := r.URL.Query()["branch"]

	if !ok || len(reqBranch) < 1 {
		log.Print("branch is missing. Assuming master")

		branch = "master"
	} else {
		branch = reqBranch[0]
	}

	log.Print("parsed branch: ", branch)

	if filematch {
		reqFiles, ok := r.URL.Query()["files"]

		if ok && len(reqFiles) > 0 {
			files = reqFiles
		}
	}

	return repo, branch, files, nil
}

func parseJSONRequest(r *http.Request, filematch bool) ([]string, string, []string, error) {
	repo := []string{}
	branch := "master"
	files := []string{}

	type gitlabProject struct {
		Gitsshurl  string `json:"git_ssh_url"`
		Githttpurl string `json:"git_http_url"`
	}

	type gitlabCommit struct {
		Added    []string
		Modified []string
		Removed  []string
	}

	type gitlabWebhook struct {
		Ref     string
		Project gitlabProject
		Commits []gitlabCommit
	}

	log.Print("parsing json request")

	var h gitlabWebhook

	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &h)
	if err != nil {
		return repo, branch, files, errors.New("bad request")
	}

	repo = append(repo, h.Project.Githttpurl)
	repo = append(repo, h.Project.Gitsshurl)

	if strings.Contains(h.Ref, "refs/heads/") {
		branch = strings.ReplaceAll(h.Ref, "refs/heads/", "")
	}

	for _, commit := range h.Commits {
		for _, file := range commit.Added {
			files = append(files, file)
		}
		for _, file := range commit.Modified {
			files = append(files, file)
		}
		for _, file := range commit.Removed {
			files = append(files, file)
		}
	}

	files = uniqueNonEmptyElementsOf(files)

	sort.Strings(files)

	return repo, branch, files, nil
}
