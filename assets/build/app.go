package main

import (
	"crypto/tls"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)
const (
	exitFail = 1
)

var (
	mapping = make(map[string][]string)
	quietPeriod int
	timeKeeper = make(map[string]*time.Timer)
)

func triggerJob(job string) bool {
	ret := false
	url := string(os.Getenv("JENKINS_URL") + "/job/" + job + "/build")

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return ret
	}

	// if user and token is defined, use it for basic auth
	if os.Getenv("JENKINS_USER") != "" {
		req.SetBasicAuth(os.Getenv("JENKINS_USER"), os.Getenv("JENKINS_TOKEN"))
	} else {
	// otherwise use the token for the direct build trigger
		url = string(url + "?token=" + os.Getenv("JENKINS_TOKEN"))
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	timeout := time.Duration(5 * time.Second)
	client := &http.Client{Transport: tr, Timeout: timeout}
	resp, err := client.Do(req)

	if err != nil {
		log.Print("Error:", err)

		return ret
	}

	if !(200 <= resp.StatusCode && resp.StatusCode <= 299) {
		log.Printf("... %v failed with status code %v\n", job, resp.StatusCode)
	} else {
		log.Printf("... %v triggered\n", job)
	}

	return true
}

func createTimer(job string) {
	if _, ok := timeKeeper[job]; ok {
		log.Print("Reseting timer for job ", job)
		timeKeeper[job].Stop()
		delete(timeKeeper, job)
	}

	log.Printf("Creating timer for job '%s' with quiet period of %d seconds", job, quietPeriod)

	timer := time.AfterFunc(time.Second*time.Duration(quietPeriod), func() {
		log.Print("Quiet period exceeded for job ", job)
		triggerJob(job)
		if _, ok := timeKeeper[job]; ok {
			log.Print("Deleting timer for job ", job)
			delete(timeKeeper, job)
		}
	})

	timeKeeper[job] = timer
	if _, ok := timeKeeper[job]; ok {
		log.Print("Timer saved in time keeper")
	}

	return
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Print("Handling new request")

	repos, ok := r.URL.Query()["repo"]

	if !ok || len(repos) < 1 {
		log.Print("Repo is missing")
		log.Print("Aborting request handling")

		return
	}

	repo := repos[0]

	log.Print("Parsed repo:", repo)

	branchs, ok := r.URL.Query()["branch"]

	var branch string
	if !ok || len(branchs) < 1 {
		log.Print("Branch is missing. Assuming master")
		branch = "master"
	} else {
		branch = branchs[0]
	}

	log.Print("Parsed branch: ", branch)

	key := BuildMappingKey([]string{repo, branch})

	log.Print("Searching mappings for key: ", key)

	if len(mapping[key]) < 1 {
		fmt.Fprintf(w, "No mappings found")
		log.Print("No mappings found")
		log.Print("Aborting request handling")
		return
	}

	log.Print("Number of mappings found: ", len(mapping[key]))

	log.Print("Start processing mappings")
	for _, job := range mapping[key] {
		createTimer(job)
	}
	log.Print("End processing mappings")

	log.Print("Handling request finished")
}

func main() {
	if err := run(os.Args, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(exitFail)
	}
}

func run(args []string, stdout io.Writer) error {
	fmt.Fprintln(stdout, "Starting trigger-proxy ...")

	fmt.Fprintln(stdout, "Checking environment variables")

	if os.Getenv("JENKINS_URL") == "" {
		return errors.New("No JENKINS_URL defined")
	}

	if os.Getenv("JENKINS_USER") == "" {
		fmt.Fprintln(stdout, "No JENKINS_USER defined")
	}

	if os.Getenv("JENKINS_TOKEN") == "" {
		return errors.New("No JENKINS_TOKEN defined")
	}

	if os.Getenv("JENKINS_MULTI") != "" {
		fmt.Fprintf(stdout, "Found multibranch project: %s\n", os.Getenv("JENKINS_MULTI"))
	}

	if os.Getenv("JENKINS_MULTI") != "" {
		os.Setenv("JENKINS_URL", os.Getenv("JENKINS_URL")+"/job/"+os.Getenv("JENKINS_MULTI"))
	}

	quietPeriod = 30
	if os.Getenv("JENKINS_QUIET") != "" {
		tQuietPeriod, err := strconv.Atoi(os.Getenv("JENKINS_QUIET"))
		if err != nil {
			return errors.New("Quiet Period could not be parsed. Aborting")
		}
		quietPeriod = tQuietPeriod
		fmt.Fprintf(stdout, "Found configured quiet period: %s\n", quietPeriod)
	}

	fmt.Fprintf(stdout, "Project URL: %s\n", os.Getenv("JENKINS_URL"))

	mappingfile := "mapping.csv"
	if os.Getenv("MAPPING_FILE") != "" {
		mappingfile = os.Getenv("MAPPING_FILE")
		fmt.Fprintf(stdout, "Found configured mapping file: %s\n", mappingfile)
	}

	if err := ParseMappingFile(mappingfile, stdout); err != nil {
		return err
	}

	http.HandleFunc("/", handler)

	fmt.Fprintln(stdout, "Serving on port 8080")
	http.ListenAndServe(":8080", nil)

	return nil
}

func ParseMappingFile(mappingfile string, stdout io.Writer) error {
        fmt.Fprintf(stdout, "Reading mapping from file: %s\n", mappingfile)

        file, err := os.Open(mappingfile)
        if err != nil {
                return err
        }
        defer file.Close()

        reader := csv.NewReader(file)
        reader.Comma = ';'
        lineCount := 0
        for {
                record, err := reader.Read()

                if err == io.EOF {
                        break
                } else if err != nil {
                        return err
                }

		key := BuildMappingKey([]string{record[0],record[1]})
                mapping[key] = append(mapping[key], record[2])
                lineCount += 1
        }

	fmt.Fprintf(stdout, "Successfully read mappings: %d\n", lineCount)

	return nil
}

func BuildMappingKey(keys []string) string {
	return strings.Join(keys, "|")
}
