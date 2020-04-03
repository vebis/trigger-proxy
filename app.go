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
	JenkinsURL   string
	JenkinsUser  string
	JenkinsToken string
	JenkinsMulti string
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
	flag.StringVar(&s.param.JenkinsURL, "jenkins-url", "", "sets the jenkins url")
	flag.StringVar(&s.param.JenkinsUser, "jenkins-user", "", "jenkins username")
	flag.StringVar(&s.param.JenkinsToken, "jenkins-token", "", "token for user or root token to trigger anonymously")
	flag.StringVar(&s.param.JenkinsMulti, "jenkins-multi", "", "root folder or job name")
	flag.StringVar(&s.param.MappingFile, "mappingfile", "mapping.csv", "path to the mapping file")
	flag.IntVar(&s.param.QuietPeriod, "quietperiod", defQp, "defines the time trigger-proxy will wait until the job is triggered")
	flag.BoolVar(&s.param.FileMatching, "filematch", false, "try to match for file names")
	flag.StringVar(&s.param.SemanticRepo, "semanticrepo", "", "repo prefix to handle as component repository")

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

	if s.param.JenkinsURL == "" {
		return errors.New("No JENKINS_URL defined")
	}

	if s.param.JenkinsUser == "" {
		log.Println("No JENKINS_USER defined")
	}

	if s.param.JenkinsToken == "" {
		return errors.New("No JENKINS_TOKEN defined")
	}

	if s.param.JenkinsMulti != "" {
		log.Printf("Found multibranch project: %s\n", s.param.JenkinsMulti)

		s.param.JenkinsURL = s.param.JenkinsURL + "/job/" + s.param.JenkinsMulti
	}

	log.Printf("Found configured quiet period: %d\n", s.param.QuietPeriod)
	log.Printf("Project URL: %s\n", s.param.JenkinsURL)

	log.Printf("Found configured mapping file: %s\n", s.param.MappingFile)

	if err := s.processMappingFile(); err != nil {
		return err
	}

	http.HandleFunc("/", s.handlePlainGet())

	log.Println("Serving on port 8080")
	http.ListenAndServe(":8080", nil)

	return nil
}
