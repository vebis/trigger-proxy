package main

import (
	"errors"
	"log"
	"net/http"
)

func parseGetRequest(r *http.Request, filematch bool) (string, string, []string, error) {
	repo := ""
	branch := ""
	files := []string{}

	log.Print("parsing get request")
	reqRepo, ok := r.URL.Query()["repo"]

	if !ok || len(reqRepo) < 1 {
		log.Print("Repo is missing")
		log.Print("Aborting request handling")

		return repo, branch, files, errors.New("repo is missing")
	}

	repo = reqRepo[0]

	log.Print("Parsed repo: ", repo)

	reqBranch, ok := r.URL.Query()["branch"]

	if !ok || len(reqBranch) < 1 {
		log.Print("Branch is missing. Assuming master")

		branch = "master"
	} else {
		branch = reqBranch[0]
	}

	log.Print("Parsed branch: ", branch)

	if filematch {
		reqFiles, ok := r.URL.Query()["file"]

		if ok && len(reqFiles) > 0 {
			files = reqFiles
		}
	}

	return repo, branch, files, nil
}
