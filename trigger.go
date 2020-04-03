package proxy

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"
)

func (s *server) triggerJob(job string) bool {
	url := createJobURL(s.param.JenkinsURL, job)

	if s.param.JenkinsUser == "" {
		url = string(url + "?token=" + s.param.JenkinsToken)
	}

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return false
	}

	// if user and token is defined, use it for basic auth
	if s.param.JenkinsUser != "" {
		req.SetBasicAuth(s.param.JenkinsUser, s.param.JenkinsToken)
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
