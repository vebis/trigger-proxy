package proxy

import (
	"errors"
	"log"
)

func (s *server) matchMappingKeys(keys []string, filematch bool) ([]string, error) {
	var hits []string
	for _, key := range keys {
		log.Print("Searching mappings for key: ", key)

		if len(s.mapping[key]) > 0 {
			for _, hit := range s.mapping[key] {
				hits = append(hits, hit)
			}
		} else if filematch {
			for len(key) > 1 {
				key = removeLastRune(key)
				if len(s.mapping[key]) > 0 {
					for _, hit := range s.mapping[key] {
						hits = append(hits, hit)
					}
				}
			}
		}
	}

	if len(hits) == 0 {
		return []string{}, errors.New("no mappings found")
	}

	log.Print("Number of mappings found: ", len(hits))

	return hits, nil
}

func (s *server) processMatching(repo, branch string, files []string, filematch bool) error {
	keys, err := evalMappingKeys(repo, branch, files, filematch, s.param.SemanticRepo)
	if err != nil {
		return err
	}

	jobs, err := s.matchMappingKeys(keys, filematch)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		s.createTimer(job)
	}

	log.Print("end processing mappings")

	return nil
}
