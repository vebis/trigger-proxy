package proxy

import (
	"errors"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	exitFail = 1
	defQp    = 10
)

type server struct {
	mapping    map[string][]string
	timeKeeper map[string]*time.Timer
	param      parameters
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
}

func main() {
	if err := run(os.Args, os.Stdout); err != nil {
		log.Fatalf("%s\n", err)
		os.Exit(exitFail)
	}
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

	flag.Parse()
}

func run(args []string, stdout io.Writer) error {
	log.Println("Starting trigger-proxy ...")

	s := server{
		mapping:    make(map[string][]string),
		timeKeeper: make(map[string]*time.Timer),
	}

	s.parseFlags(args)

	log.Println("Checking environment variables")

	if s.param.jenkins.URL == "" {
		return errors.New("No JENKINS_URL defined")
	}

	if s.param.jenkins.User == "" {
		log.Println("No JENKINS_USER defined")
	}

	if s.param.jenkins.Token == "" {
		return errors.New("No JENKINS_TOKEN defined")
	}

	if s.param.jenkins.Multi != "" {
		log.Printf("Found multibranch project: %s\n", s.param.jenkins.Multi)

		s.param.jenkins.URL = s.param.jenkins.URL + "/job/" + s.param.jenkins.Multi
	}

	log.Printf("Found configured quiet period: %d\n", s.param.proxy.QuietPeriod)
	log.Printf("Project URL: %s\n", s.param.jenkins.URL)

	log.Printf("Found configured mapping file: %s\n", s.param.proxy.MappingFile)

	if err := s.processMappingFile(); err != nil {
		return err
	}

	http.HandleFunc("/", s.handlePlainGet())

	log.Println("Serving on port 8080")
	http.ListenAndServe(":8080", nil)

	return nil
}
