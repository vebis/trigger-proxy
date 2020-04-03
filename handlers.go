package proxy

import (
	"log"
	"net/http"
)

func (s *server) handlePlainGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("handling new request")

		repo, branch, files, err := parseGetRequest(r, s.param.FileMatching)

		if err != nil {
			log.Print(err)
			log.Print("aborting request handling")

			w.WriteHeader(http.StatusBadRequest)

			return
		}

		if err := s.processMatching(repo, branch, files); err != nil {
			log.Print(err)
			http.NotFound(w, r)

			return
		}

		w.WriteHeader(http.StatusOK)

		log.Print("handling of request finished")
	}
}
