package main

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"os"
)

func (s *server) refreshMapping() error {
	newHash, err := s.hashMappingFile()
	if err != nil {
		return err
	}

	if s.mappingHash != "" {
		log.Printf("hash of mapping file has changed (old: %s, new %s)", s.mappingHash, newHash)
	}

	if newHash != s.mappingHash {
		if err := s.processMappingFile(); err != nil {
			return err
		}
	}

	return nil
}

// processMappingFile processes the file at given path
func (s *server) processMappingFile() error {
	log.Printf("Reading mapping from file: %s\n", s.param.proxy.MappingFile)

	file, err := os.Open(s.param.proxy.MappingFile)
	if err != nil {
		return err
	}
	defer file.Close()

	mapping, perr := parseMappingFile(file, s.param.proxy.FileMatching)
	if perr != nil {
		return perr
	}
	newHash, herr := s.hashMappingFile()
	if herr != nil {
		return herr
	}
	s.mapping = mapping
	s.mappingHash = newHash

	return nil
}

// parseMappingFile parses the given file and returns the mapping
func parseMappingFile(file io.Reader, filematch bool) (map[string][]string, error) {
	var m = make(map[string][]string)

	reader := csv.NewReader(file)
	reader.Comma = ','
	lineCount := 0
	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			return m, err
		}

		var key string
		if filematch {
			if len(record) != 4 {
				return m, errors.New("no file matching information provided in mapping file")
			}
			key = buildMappingKey([]string{record[0], record[1], record[3]})
		} else {
			key = buildMappingKey([]string{record[0], record[1]})
		}
		m[key] = append(m[key], record[2])
		lineCount++
	}

	log.Printf("Successfully read mappings: %d\n", lineCount)

	return m, nil
}

func (s *server) hashMappingFile() (string, error) {
	var mhash string
	file, err := os.Open(s.param.proxy.MappingFile)
	if err != nil {
		return mhash, err
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return mhash, err
	}

	mhash = hex.EncodeToString(h.Sum(nil))

	return mhash, nil
}
