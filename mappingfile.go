package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func (s *server) refreshMapping() error {
	var newHash string
	if s.isMappingURL() {
		nash, err := s.hashMappingURL()
		if err != nil {
			return err
		}
		newHash = nash
	} else {
		nash, err := s.hashMappingFile()
		if err != nil {
			return err
		}
		newHash = nash
	}

	if newHash != s.mappingHash {
		if s.mappingHash != "" {
			log.Printf("hash of mapping has changed (old: %s, new %s)", s.mappingHash, newHash)
		}
		if s.isMappingURL() {
			if err := s.processMappingURL(); err != nil {
				return err
			}
		} else {
			if err := s.processMappingFile(); err != nil {
				return err
			}
		}
	}

	return nil
}

// processMappingFile processes the file at given path
func (s *server) processMappingFile() error {
	log.Printf("reading mapping from file: %s\n", s.param.proxy.Mapping.file)

	file, err := os.Open(s.param.proxy.Mapping.file)
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

func (s *server) processMappingURL() error {
	log.Printf("reading mapping from url: %s\n", s.param.proxy.Mapping.url)

	body, err := httpGetWrapper(s.param.proxy.Mapping.url)
	if err != nil {
		return err
	}

	mapping, err := parseMappingFile(bytes.NewReader(body), s.param.proxy.FileMatching)
	if err != nil {
		return err
	}
	newHash, err := s.hashMappingURL()
	if err != nil {
		return err
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
	file, err := os.Open(s.param.proxy.Mapping.file)
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

func (s *server) hashMappingURL() (string, error) {
	var mhash string
	body, err := httpGetWrapper(s.param.proxy.Mapping.url + ".sha256")
	if err != nil {
		return mhash, err
	}

	mhash = string(body)

	return mhash, err
}

func (s *server) isMappingURL() bool {
	if len(s.param.proxy.Mapping.url) > 0 {
		return true
	}

	return false
}

func httpGetWrapper(url string) ([]byte, error) {
	var rbody []byte

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return rbody, err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	timeout := time.Duration(5 * time.Second)
	client := &http.Client{Transport: tr, Timeout: timeout}
	resp, err := client.Do(req)

	if err != nil {
		return rbody, err
	}
	defer resp.Body.Close()

	if !(200 <= resp.StatusCode && resp.StatusCode <= 299) {
		return rbody, errors.New("failure in getting a succesful response")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return rbody, err
	}

	rbody = body

	return rbody, nil
}
