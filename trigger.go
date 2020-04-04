package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"
)

func (s *server) triggerJob(job string) bool {
	url := createJobURL(s.param.jenkins.URL, job)

	if s.param.jenkins.User == "" {
		url = string(url + "?token=" + s.param.jenkins.Token)
	}

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return false
	}

	// if user and token is defined, use it for basic auth
	if s.param.jenkins.User != "" {
		req.SetBasicAuth(s.param.jenkins.User, s.param.jenkins.Token)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	timeout := time.Duration(5 * time.Second)
	client := &http.Client{Transport: tr, Timeout: timeout}
	resp, err := client.Do(req)

	if err != nil {
		log.Print("Error:", err)

		return false
	}

	if !(200 <= resp.StatusCode && resp.StatusCode <= 299) {
		log.Printf("... %v failed with status code %v\n", job, resp.StatusCode)
	} else {
		log.Printf("... %v triggered\n", job)
	}

	return true
}
