package main

import (
	"encoding/csv"
	"errors"
	"io"
	"log"
	"os"
)

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

	s.mapping = mapping

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
