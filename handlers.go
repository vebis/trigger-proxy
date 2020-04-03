package proxy

import (
	"log"
	"net/http"
)

func (s *server) handlePlainGet(w http.ResponseWriter, r *http.Request) {
	log.Print("handling new request")

	repo, branch, files, err := parseGetRequest(r, s.param.FileMatching)

	if err != nil {
		log.Print(err)
		log.Print("aborting request handling")

		return
	}

	if err := s.processMatching(repo, branch, files, s.param.FileMatching); err != nil {
		log.Print(err)
		http.NotFound(w, r)

		return
	}

	log.Print("handling of request finished")
}
