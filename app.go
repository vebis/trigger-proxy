package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	exitFail = 1
	defQp    = 10
	defPort  = 8080
	defInt   = 5
)

type server struct {
	mapping                map[string][]string
	mappingHash            string
	mappingRefreshInterval time.Duration
	timeKeeper             map[string]*time.Timer
	param                  parameters
}

type parameters struct {
	jenkins jenkins
	proxy   proxy
}

type jenkins struct {
	URL   string
	User  string
	Token string
	Multi string
}

type proxy struct {
	MappingFile  string
	QuietPeriod  int
	FileMatching bool
	SemanticRepo string
	port         int
}

func main() {
	if err := run(os.Args); err != nil {
		log.Fatalf("%s\n", err)
		os.Exit(exitFail)
	}
}

// newServer returns a new trigger proxy server
func newServer(args []string) (server, error) {
	s := server{
		mapping:     make(map[string][]string),
		mappingHash: "",
		timeKeeper:  make(map[string]*time.Timer),
	}

	s.parseFlags(args)

	log.Println("checking configuration")

	if s.param.jenkins.URL == "" {
		return s, errors.New("no jenkins url defined")
	}

	if s.param.jenkins.Token == "" {
		return s, errors.New("no jenkins token defined")
	}

	log.Printf("project url: %s\n", s.param.jenkins.URL)

	if s.param.jenkins.User == "" {
		log.Println("no jenkins user defined")
	} else {
		log.Printf("jenkins user: %s\n", s.param.jenkins.User)
	}

	if s.param.jenkins.Multi != "" {
		log.Printf("found multibranch project: %s\n", s.param.jenkins.Multi)

		s.param.jenkins.URL = s.param.jenkins.URL + "/job/" + s.param.jenkins.Multi
	}

	log.Printf("quiet period: %d\n", s.param.proxy.QuietPeriod)
	log.Printf("mapping file: %s\n", s.param.proxy.MappingFile)

	if s.param.proxy.SemanticRepo != "" {
		s.param.proxy.FileMatching = true
	}

	return s, nil
}
func (s *server) parseFlags(args []string) {
	flag.StringVar(&s.param.jenkins.URL, "jenkins-url", "", "sets the jenkins url")
	flag.StringVar(&s.param.jenkins.User, "jenkins-user", "", "jenkins username")
	flag.StringVar(&s.param.jenkins.Token, "jenkins-token", "", "token for user or root token to trigger anonymously")
	flag.StringVar(&s.param.jenkins.Multi, "jenkins-multi", "", "root folder or job name")
	flag.StringVar(&s.param.proxy.MappingFile, "mappingfile", "mapping.csv", "path to the mapping file")
	flag.IntVar(&s.param.proxy.QuietPeriod, "quietperiod", defQp, "defines the time trigger-proxy will wait until the job is triggered")
	flag.BoolVar(&s.param.proxy.FileMatching, "filematch", false, "try to match for file names")
	flag.StringVar(&s.param.proxy.SemanticRepo, "semanticrepo", "", "repo prefix to handle as component repository")
	flag.IntVar(&s.param.proxy.port, "port", defPort, "defines the http port to listen on")

	refreshInterval := flag.Int("mappingrefresh", defInt, "refresh interval in minutes to check for modified mapping file")
	s.mappingRefreshInterval = time.Duration(*refreshInterval) * time.Minute

	flag.Parse()
}

func run(args []string) error {
	log.Println("Starting trigger-proxy ...")

	s, err := newServer(args)
	if err != nil {
		return err
	}

	if err := s.refreshMapping(); err != nil {
		return err
	}

	s.createRefreshJob()

	http.HandleFunc("/", s.handlePlainGet())
	http.HandleFunc("/json", s.handleJSONPost())

	port := strconv.Itoa(s.param.proxy.port)
	log.Println("Serving on port " + port)
	http.ListenAndServe(":"+port, nil)

	return nil
}
