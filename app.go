package main

import (
    "fmt"
    "net/http"
    "encoding/csv"
    "io"
    "os"
    "crypto/tls"
    "time"
)

var mapping = make(map[string][]string)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Handling new request")

    repos, ok := r.URL.Query()["repo"]

    if !ok || len(repos) < 1 {
        fmt.Println("Repo is missing")
        fmt.Println("Aborting request handling")
        return
    }

    repo := repos[0]

    fmt.Println("Parsed repo:", repo)

    branchs, ok := r.URL.Query()["branch"]

    var branch string
    if !ok || len(branchs) < 1 {
        fmt.Println("Branch is missing. Assuming master")
	branch = "master"
    } else {
        branch = branchs[0]
    }

    fmt.Println("Parsed branch:", branch)

    key := string(repo+"|"+branch)

    fmt.Println("Searching mappings for key:", key)

    if len(mapping[key]) < 1 {
	fmt.Fprintf(w, "No mappings found")
        fmt.Println("No mappings found")
	fmt.Println("Aborting request handling")
	return
    } else {
	fmt.Println("Number of mappings found:", len(mapping[key]))
    }

    fmt.Println("Start processing mappings")
    for _, job := range mapping[key] {
	url := string(os.Getenv("JENKINS_URL")+"/job/"+job+"/build")

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
		fmt.Println("Error:", err)

		return
	}

	if !(200 <= resp.StatusCode && resp.StatusCode <= 299) {
		fmt.Printf("... %v failed with status code %v\n", job, resp.StatusCode)
	} else {
		fmt.Printf("... %v triggered\n", job)
	}
    }
    fmt.Println("End processing mappings")

    fmt.Println("Handling request finished")
}

func main() {
    fmt.Println("Starting trigger-proxy ...")

    fmt.Println("Checking environment variables")

    if os.Getenv("JENKINS_URL") == "" {
	fmt.Println("No JENKINS_URL defined")
	return
    }

    if os.Getenv("JENKINS_USER") == "" {
        fmt.Println("No JENKINS_USER defined")
        return
    }

    if os.Getenv("JENKINS_TOKEN") == "" {
        fmt.Println("No JENKINS_TOKEN defined")
        return
    }

    if os.Getenv("JENKINS_MULTI") != "" {
        fmt.Println("Found multibranch project:", os.Getenv("JENKINS_MULTI"))
    }

    if os.Getenv("JENKINS_MULTI") != "" {
        os.Setenv("JENKINS_URL", os.Getenv("JENKINS_URL")+"/job/"+os.Getenv("JENKINS_MULTI"))
    }

    fmt.Println("Project URL:", os.Getenv("JENKINS_URL"))

    mappingfile := "mapping.csv"
    if os.Getenv("MAPPING_FILE") != "" {
        mappingfile = os.Getenv("MAPPING_FILE")
	fmt.Println("Found configured mapping file:", mappingfile)
    }

    fmt.Println("Reading mapping from file:", mappingfile)

    file, err := os.Open(mappingfile)
	if err != nil {
		// err is printable
		// elements passed are separated by space automatically
		fmt.Println("Error:", err)
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
			fmt.Println("Error:", err)
			return
		}

		mapping[record[0]+"|"+record[1]] = append(mapping[record[0]+"|"+record[1]], record[2])
		lineCount += 1
	}

    fmt.Println("Succesfully read mappings:", lineCount)

    fmt.Println("Adding handler")
    http.HandleFunc("/", handler)

    fmt.Println("Serving on port 8080")
    http.ListenAndServe(":8080", nil)
}
