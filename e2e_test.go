package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"testing"
	"time"
)

func Test_e2e_mapping_file(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantStatus int
		wantErr    bool
	}{
		{
			"simple_match",
			[]string{"proxy",
				"-jenkins-url=http://localhost:8081/",
				"-jenkins-token=whatever",
				"-mapping-file=./examples/example.csv",
				"-quietperiod=1",
				"-port=8080",
			},
			http.StatusOK,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// send trigger-proxy into background
			go func() {
				err := run(tt.args)
				if (err != nil) != tt.wantErr {
					t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
				}
			}()
			// wait for initialization and listening to requests
			for {
				conn, _ := net.DialTimeout("tcp", net.JoinHostPort("", "8080"), time.Millisecond*time.Duration(10))
				if conn != nil {
					conn.Close()
					break
				}
			}
			// setup mockup jenkins
			go func() {
				mockHandler := func(w http.ResponseWriter, req *http.Request) {
					log.Printf("%v", req.URL.Query())
					io.WriteString(w, "Hello, world!\n")
				}

				http.HandleFunc("/job/job1/build", mockHandler)
				log.Fatal(http.ListenAndServe(":8081", nil))
			}()
			// wait for initialization and listening to requests of mockup server
			for {
				conn, _ := net.DialTimeout("tcp", net.JoinHostPort("", "8081"), time.Millisecond*time.Duration(10))
				if conn != nil {
					conn.Close()
					break
				}
			}
			// send request to trigger-proxy
			_, err := http.Get("http://localhost:8080/?repo=git://gitserver/git/testrepo1&branch=master")
			if (err != nil) != tt.wantErr {
				t.Errorf("request creation error = %v, wantErr %v", err, tt.wantErr)
			}
			time.Sleep(time.Second * time.Duration(2))
		})
	}
}

func Test_e2e_mapping_url(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantStatus int
		wantErr    bool
	}{
		{
			"simple_match",
			[]string{"proxy",
				"-jenkins-url=http://localhost:8081/",
				"-jenkins-token=whatever",
				"-mapping-url=http://localhost:8083/example.csv",
				"-quietperiod=1",
				"-port=8080",
			},
			http.StatusOK,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go func() {
				http.HandleFunc("/example.csv", staticServeHandler("example.csv"))
				http.HandleFunc("/example.csv.sha265", staticServeHandler("example.csv.sha256"))
				log.Fatal(http.ListenAndServe(":8083", nil))
			}()
			// wait for initialization and listening to requests
			for {
				conn, _ := net.DialTimeout("tcp", net.JoinHostPort("", "8083"), time.Millisecond*time.Duration(10))
				if conn != nil {
					conn.Close()
					break
				}
			}

			// send trigger-proxy into background
			go func() {
				err := run(tt.args)
				if (err != nil) != tt.wantErr {
					t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
				}
			}()
			// wait for initialization and listening to requests
			for {
				conn, _ := net.DialTimeout("tcp", net.JoinHostPort("", "8080"), time.Millisecond*time.Duration(10))
				if conn != nil {
					conn.Close()
					break
				}
			}
			// setup mockup jenkins
			go func() {
				mockHandler := func(w http.ResponseWriter, req *http.Request) {
					log.Printf("%v", req.URL.Query())
				}

				http.HandleFunc("/job/job2/build", mockHandler)
				log.Fatal(http.ListenAndServe(":8081", nil))
			}()
			// wait for initialization and listening to requests of mockup server
			for {
				conn, _ := net.DialTimeout("tcp", net.JoinHostPort("", "8081"), time.Millisecond*time.Duration(10))
				if conn != nil {
					conn.Close()
					break
				}
			}
			// send request to trigger-proxy
			_, err := http.Get("http://localhost:8080/?repo=git://gitserver/git/testrepo3&branch=master")
			if (err != nil) != tt.wantErr {
				t.Errorf("request creation error = %v, wantErr %v", err, tt.wantErr)
			}
			time.Sleep(time.Second * time.Duration(2))
		})
	}
}

func staticServeHandler(filename string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, _ := os.Open("./examples/" + filename)
		defer file.Close()

		io.Copy(w, file)

		return
	}
}
