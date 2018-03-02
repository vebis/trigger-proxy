package main

import (
	"crypto/tls"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var mapping = make(map[string][]string)
var quietPeriod int
var time_keeper = make(map[string]*time.Timer)

func triggerJob(job string) bool {
	ret := false
	url := string(os.Getenv("JENKINS_URL") + "/job/" + job + "/build")

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return ret
	}

	req.SetBasicAuth(os.Getenv("JENKINS_USER"), os.Getenv("JENKINS_TOKEN"))

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
	if _, ok := time_keeper[job]; ok {
		log.Print("Reseting timer for job ", job)
		time_keeper[job].Stop()
		delete(time_keeper, job)
	}

	log.Printf("Creating timer for job '%s' with quiet period of %s seconds", job, quietPeriod)

	timer := time.AfterFunc(time.Second*time.Duration(quietPeriod), func() {
		log.Print("Quiet period exceeded for job ", job)
		triggerJob(job)
		if _, ok := time_keeper[job]; ok {
			log.Print("Deleting timer for job ", job)
			delete(time_keeper, job)
		}
	})
	//    defer timer.Stop()

	time_keeper[job] = timer
	if _, ok := time_keeper[job]; ok {
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

	log.Print("Parsed branch:", branch)

	key := string(repo + "|" + branch)

	log.Print("Searching mappings for key:", key)

	if len(mapping[key]) < 1 {
		fmt.Fprintf(w, "No mappings found")
		log.Print("No mappings found")
		log.Print("Aborting request handling")
		return
	}

	log.Print("Number of mappings found:", len(mapping[key]))

	log.Print("Start processing mappings")
	for _, job := range mapping[key] {
		createTimer(job)
	}
	log.Print("End processing mappings")

	log.Print("Handling request finished")
}

func main() {
	log.Print("Starting trigger-proxy ...")

	log.Print("Checking environment variables")

	if os.Getenv("JENKINS_URL") == "" {
		log.Print("No JENKINS_URL defined")
		return
	}

	if os.Getenv("JENKINS_USER") == "" {
		log.Print("No JENKINS_USER defined")
		return
	}

	if os.Getenv("JENKINS_TOKEN") == "" {
		log.Print("No JENKINS_TOKEN defined")
		return
	}

	if os.Getenv("JENKINS_MULTI") != "" {
		log.Print("Found multibranch project:", os.Getenv("JENKINS_MULTI"))
	}

	if os.Getenv("JENKINS_MULTI") != "" {
		os.Setenv("JENKINS_URL", os.Getenv("JENKINS_URL")+"/job/"+os.Getenv("JENKINS_MULTI"))
	}

	quietPeriod = 30
	if os.Getenv("JENKINS_QUIET") != "" {
		tQuietPeriod, err := strconv.Atoi(os.Getenv("JENKINS_QUIET"))
		if err != nil {
			log.Fatal("Quiet Period could not be parsed. Aborting")
		}
		quietPeriod = tQuietPeriod
		log.Print("Found configured quiet period:", quietPeriod)
	}

	log.Print("Project URL:", os.Getenv("JENKINS_URL"))

	mappingfile := "mapping.csv"
	if os.Getenv("MAPPING_FILE") != "" {
		mappingfile = os.Getenv("MAPPING_FILE")
		log.Print("Found configured mapping file:", mappingfile)
	}

	log.Print("Reading mapping from file:", mappingfile)

	file, err := os.Open(mappingfile)
	if err != nil {
		// err is printable
		// elements passed are separated by space automatically
		log.Print("Error:", err)
		return
	}
	// automatically call Close() at the end of current method
	defer file.Close()
	//

	reader := csv.NewReader(file)
	// options are available at:
	// http://golang.org/src/pkg/encoding/csv/reader.go?s=3213:3671#L94
	reader.Comma = ';'
	lineCount := 0
	for {
		// read just one record, but we could ReadAll() as well
		record, err := reader.Read()
		// end-of-file is fitted into err
		if err == io.EOF {
			break
		} else if err != nil {
			log.Print("Error:", err)
			return
		}

		mapping[record[0]+"|"+record[1]] = append(mapping[record[0]+"|"+record[1]], record[2])
		lineCount += 1
	}

	log.Print("Succesfully read mappings:", lineCount)

	log.Print("Adding handler")
	http.HandleFunc("/", handler)

	log.Print("Serving on port 8080")
	http.ListenAndServe(":8080", nil)
}
