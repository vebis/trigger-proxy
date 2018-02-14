package main

import (
    "fmt"
    "net/http"
    "encoding/csv"
    "io"
    "os"
    "crypto/tls"
    "time"
    "log"
)

var mapping = make(map[string][]string)
var quiet_period string

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

    key := string(repo+"|"+branch)

    log.Print("Searching mappings for key:", key)

    if len(mapping[key]) < 1 {
        fmt.Fprintf(w, "No mappings found")
        log.Print("No mappings found")
        log.Print("Aborting request handling")
        return
    } else {
        log.Print("Number of mappings found:", len(mapping[key]))
    }

    log.Print("Start processing mappings")
    for _, job := range mapping[key] {
        url := string(os.Getenv("JENKINS_URL")+"/job/"+job+"/build?delay="+quiet_period)

        req, err := http.NewRequest("POST", url, nil)
        if err != nil {
            return
        }

            req.SetBasicAuth(os.Getenv("JENKINS_USER"), os.Getenv("JENKINS_TOKEN"))

        tr := &http.Transport{
                TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
        }

        timeout := time.Duration(5 * time.Second)
        client := &http.Client{Transport: tr, Timeout: timeout}
        resp, err := client.Do(req)

        if err != nil {
            fmt.Fprintf(w, "some error occured, check log")
            log.Print("Error:", err)

            return
        }

        if !(200 <= resp.StatusCode && resp.StatusCode <= 299) {
            fmt.Printf("... %v failed with status code %v\n", job, resp.StatusCode)
        } else {
            fmt.Printf("... %v triggered\n", job)
        }
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

    quiet_period = "30"
    if os.Getenv("JENKINS_QUIET") != "" {
        quiet_period = os.Getenv("JENKINS_QUIET")
        log.Print("Found configured quiet period:", quiet_period)
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
