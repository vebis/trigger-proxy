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
	defQp    = 10   // default quiet period (in sec)
	defPort  = 8080 // default http port
	defInt   = 5    // default interfall of mapping refresh (in min)
)

type mapping map[string][]string

type server struct {
	mapping                mapping
	mappingHash            string
	mappingSource          mappingHandler
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

type mappingSource struct {
	path string
	hash string
}

type proxy struct {
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
		mapping:     make(mapping),
		mappingHash: "",
		timeKeeper:  make(map[string]*time.Timer),
	}

	if err := s.parseFlags(args); err != nil {
		return s, err
	}

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

	// log.Printf("mapping source: %s\n", s.mappingSource)

	log.Println("---------------------------------")

	if s.param.proxy.SemanticRepo != "" {
		s.param.proxy.FileMatching = true
	}

	return s, nil
}

func (s *server) parseFlags(args []string) error {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)

	flags.StringVar(&s.param.jenkins.URL, "jenkins-url", "", "sets the jenkins url")
	flags.StringVar(&s.param.jenkins.User, "jenkins-user", "", "jenkins username")
	flags.StringVar(&s.param.jenkins.Token, "jenkins-token", "", "token for user or root token to trigger anonymously")
	flags.StringVar(&s.param.jenkins.Multi, "jenkins-multi", "", "root folder or job name")

	flags.IntVar(&s.param.proxy.QuietPeriod, "quietperiod", defQp, "defines the time trigger-proxy will wait until the job is triggered")
	flags.BoolVar(&s.param.proxy.FileMatching, "filematch", false, "try to match for file names")
	flags.StringVar(&s.param.proxy.SemanticRepo, "semanticrepo", "", "repo prefix to handle as component repository")
	flags.IntVar(&s.param.proxy.port, "port", defPort, "defines the http port to listen on")

	refreshInterval := flags.Int("mappingrefresh", defInt, "refresh interval in minutes to check for modified mapping file")
	s.mappingRefreshInterval = time.Duration(*refreshInterval) * time.Minute

	var (
		mFile string
		mURL  string
	)
	flags.StringVar(&mFile, "mapping-file", "mapping.csv", "path to the mapping file")
	flags.StringVar(&mURL, "mapping-url", "", "path to the mapping file")

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	// if an URL is defined, use that
	if len(mURL) > 0 {
		s.mappingSource = mappingURL{
			path: mURL,
		}
	} else {
		if len(mFile) == 0 {
			return errors.New("no valid mapping source specified")
		}

		s.mappingSource = mappingFile{
			path: mFile,
		}
	}

	return nil
}

func run(args []string) error {
	log.Println("starting trigger-proxy ...")

	s, err := newServer(args)
	if err != nil {
		return err
	}

	// if the first refresh fails, fail early
	if err := s.refreshMapping(); err != nil {
		return err
	}

	s.createRefreshJob()

	http.HandleFunc("/", s.handlePlainGet())
	http.HandleFunc("/json", s.handleJSONPost())
	http.HandleFunc("/readyz", s.handleReadiness())

	port := strconv.Itoa(s.param.proxy.port)
	log.Println("serving on port " + port)
	http.ListenAndServe(":"+port, nil)

	return nil
}
