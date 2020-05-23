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

type mappingHandler interface {
	hashSource() (string, error)
	process(bool) (mapping, string, error)
}

type mappingFile mappingSource
type mappingURL mappingSource

func (s *server) refreshMapping() error {
	newHash, err := s.mappingSource.hashSource()
	if err != nil {
		return err
	}

	if newHash != s.mappingHash {
		if s.mappingHash != "" {
			log.Printf("hash of mapping has changed (old: %s, new: %s)", s.mappingHash, newHash)
		}

		curMapping, curHash, err := s.mappingSource.process(s.param.proxy.FileMatching)

		if err != nil {
			return err
		}
		s.mapping = curMapping
		s.mappingHash = curHash
	}

	return nil
}

// processMappingFile processes the file at given path
func (m mappingFile) process(fileMatching bool) (mapping, string, error) {
	log.Printf("reading mapping from file: %s\n", m.path)
	var (
		nm mapping
		nh string
	)

	file, err := os.Open(m.path)
	if err != nil {
		return nm, nh, err
	}
	defer file.Close()

	mapping, perr := parseMappingFile(file, fileMatching)
	if perr != nil {
		return nm, nh, perr
	}
	newHash, herr := m.hashSource()
	if herr != nil {
		return nm, nh, herr
	}

	nm = mapping
	nh = newHash

	return nm, nh, nil
}

func (m mappingURL) process(fileMatching bool) (mapping, string, error) {
	log.Printf("reading mapping from url: %s\n", m.path)
	var (
		nm mapping
		nh string
	)
	body, err := httpGetWrapper(m.path)
	if err != nil {
		return nm, nh, err
	}

	mapping, err := parseMappingFile(bytes.NewReader(body), fileMatching)
	if err != nil {
		return nm, nh, err
	}
	newHash, err := m.hashSource()
	if err != nil {
		return nm, nh, err
	}
	nm = mapping
	nh = newHash

	return nm, nh, nil
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

	log.Printf("successfully read mappings: %d\n", lineCount)

	return m, nil
}

func (m mappingFile) hashSource() (string, error) {
	var mhash string
	file, err := os.Open(m.path)
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

func (m mappingURL) hashSource() (string, error) {
	var mhash string
	body, err := httpGetWrapper(m.path + ".sha256")
	if err != nil {
		return mhash, err
	}

	mhash = string(body)

	return mhash, err
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
