package main

import (
	"errors"
	"log"
	"strings"
)

func buildMappingKey(keys []string) string {
	return strings.Join(keys, "|")
}

func createJobURL(jenkinsURL, job string) string {
	return string(jenkinsURL + "/job/" + job + "/build")
}

func removeLastRune(s string) string {
	if len(s) <= 1 {
		return ""
	}
	r := []rune(s)
	return string(r[:len(r)-1])
}

func appendMappingKeys(keys, files []string, repo, branch, fileprefix string) []string {
	for _, file := range files {
		keys = append(keys, buildMappingKey([]string{repo, branch, fileprefix + file}))
	}
	return keys
}

func evalMappingKeys(repo, branch string, files []string, filematch bool, sematicrepo string) []string {
	var keys []string
	if filematch {
		if sematicrepo != "" {
			pkgname, err := getSemanticRepoName(repo, sematicrepo)
			if err == nil && pkgname != "" {
				keys = appendMappingKeys(keys, files, repo, branch, pkgname+"/")
			} else {
				keys = appendMappingKeys(keys, files, repo, branch, "")
			}
		} else {
			keys = appendMappingKeys(keys, files, repo, branch, "")
		}
	} else {
		key := buildMappingKey([]string{repo, branch})

		keys = append(keys, key)
	}

	return keys
}

func getSemanticRepoName(repo, semanticRepo string) (string, error) {
	reponame := strings.ReplaceAll(repo, ".git", "")
	reponame = strings.ReplaceAll(reponame, semanticRepo, "")

	if strings.ContainsAny(reponame, "/:.") {
		return "", errors.New("semantic repo reduction failed")
	}

	log.Printf("found semantic repo: %s\n", reponame)

	return reponame, nil
}

func uniqueNonEmptyElementsOf(s []string) []string {
	unique := make(map[string]bool, len(s))
	us := make([]string, len(unique))
	for _, elem := range s {
		if len(elem) != 0 {
			if !unique[elem] {
				us = append(us, elem)
				unique[elem] = true
			}
		}
	}

	return us

}
