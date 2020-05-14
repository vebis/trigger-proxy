package main

import (
	"errors"
	"log"
)

func (s *server) getHits(hits []string, key string) []string {
	if len(s.mapping[key]) > 0 {
		for _, hit := range s.mapping[key] {
			hits = append(hits, hit)
		}
	}

	return hits
}

func (s *server) matchMappingKeys(keys []string, filematch bool) ([]string, error) {
	var hits []string
	for _, key := range keys {
		log.Print("searching mappings for key: ", key)

		if len(s.mapping[key]) > 0 {
			hits = s.getHits(hits, key)
		} else if filematch {
			for len(key) > 1 {
				key = removeLastRune(key)
				oldHitCount := len(hits)
				hits = s.getHits(hits, key)
				if len(hits) > oldHitCount {
					break
				}
			}
		}
	}

	if len(hits) == 0 {
		return []string{}, errors.New("no mappings found")
	}

	log.Print("number of mappings found: ", len(hits))

	return hits, nil
}

func (s *server) processMatching(repo, branch string, files []string) error {
	keys := evalMappingKeys(repo, branch, files, s.param.proxy.FileMatching, s.param.proxy.SemanticRepo)

	jobs, err := s.matchMappingKeys(keys, s.param.proxy.FileMatching)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		s.createTimer(job)
	}

	log.Print("end processing mappings")

	return nil
}
