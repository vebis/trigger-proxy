package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func Test_parseMappingFile(t *testing.T) {
	type args struct {
		file      io.Reader
		filematch bool
	}
	tests := []struct {
		name    string
		args    args
		want    map[string][]string
		wantErr bool
	}{
		{
			"single_repo",
			args{file: strings.NewReader("git://repo/repo,branch,job,"), filematch: false},
			map[string][]string{
				"git://repo/repo|branch": {"job"},
			},
			false,
		},
		{
			"single_repo_filematch",
			args{file: strings.NewReader("git://repo/repo,branch,job,repo"), filematch: true},
			map[string][]string{
				"git://repo/repo|branch|repo": {"job"},
			},
			false,
		},
		{
			"single_repo_filematch_fail",
			args{file: strings.NewReader("git://repo/repo,branch,job"), filematch: true},
			map[string][]string{},
			true,
		},
		{
			"three_repos",
			args{file: strings.NewReader("git://repo/repo,branch,job\ngit://repo/repo,branch,job2\ngit://repo/repo2,branch,job"), filematch: false},
			map[string][]string{
				"git://repo/repo|branch":  {"job", "job2"},
				"git://repo/repo2|branch": {"job"},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMappingFile(tt.args.file, tt.args.filematch)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMappingFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseMappingFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_server_processMappingFile(t *testing.T) {
	tests := []struct {
		name    string
		s       server
		want    int
		wantErr bool
	}{
		{
			"example_parser_nofilematch",
			server{
				param: parameters{
					proxy: proxy{
						Mapping: mapping{
							file: "./examples/example.csv",
						},
					},
				},
			},
			8,
			false,
		},
		{
			"example_parser_filematch",
			server{
				param: parameters{
					proxy: proxy{
						Mapping: mapping{
							file: "./examples/example_fm.csv",
						},
						FileMatching: true,
					},
				},
			},
			2,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.processMappingFile(); (err != nil) != tt.wantErr {
				t.Errorf("server.processMappingFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			got := 0
			for k := range tt.s.mapping {
				for range tt.s.mapping[k] {
					got = got + 1
				}
			}
			if got != tt.want {
				t.Errorf("server.processMappingFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_server_refreshMapping(t *testing.T) {
	tests := []struct {
		name    string
		s       server
		want    int
		wantErr bool
	}{
		{
			"example_parser_nofilematch",
			server{
				param: parameters{
					proxy: proxy{
						Mapping: mapping{
							file: "./examples/example.csv",
						},
					},
				},
			},
			8,
			false,
		},
		{
			"example_parser_filematch",
			server{
				param: parameters{
					proxy: proxy{
						Mapping: mapping{
							file: "./examples/example_fm.csv",
						},
						FileMatching: true,
					},
				},
			},
			2,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.refreshMapping(); (err != nil) != tt.wantErr {
				t.Errorf("server.refreshMapping() error = %v, wantErr %v", err, tt.wantErr)
			}
			got := 0
			for k := range tt.s.mapping {
				for range tt.s.mapping[k] {
					got = got + 1
				}
			}
			if got != tt.want {
				t.Errorf("server.refreshMapping() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_server_advanced_refreshMapping(t *testing.T) {
	tests := []struct {
		name    string
		s       server
		newMap  string
		want    int
		wantErr bool
	}{
		{
			"refresh_mapping",
			server{
				param: parameters{
					proxy: proxy{
						Mapping: mapping{
							file: "./examples/example.csv",
						},
					},
				},
			},
			"./examples/example_refresh.csv",
			7,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.refreshMapping(); (err != nil) != tt.wantErr {
				t.Errorf("server.refreshMapping() error = %v, wantErr %v", err, tt.wantErr)
			}

			tt.s.param.proxy.Mapping.file = tt.newMap

			if err := tt.s.refreshMapping(); (err != nil) != tt.wantErr {
				t.Errorf("server.refreshMapping() error = %v, wantErr %v", err, tt.wantErr)
			}
			got := 0
			for k := range tt.s.mapping {
				for range tt.s.mapping[k] {
					got = got + 1
				}
			}
			if got != tt.want {
				t.Errorf("server.refreshMapping() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_server_advanced_refreshMapping_url(t *testing.T) {
	tests := []struct {
		name    string
		s       server
		newMap  string
		want    int
		wantErr bool
	}{
		{
			"refresh_mapping",
			server{
				param: parameters{
					proxy: proxy{
						Mapping: mapping{
							url: "http://localhost:10000/example.csv",
						},
					},
				},
			},
			"http://localhost:10000/example_refresh.csv",
			7,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go func() {
				http.HandleFunc("/example.csv", staticServeHandler("example.csv"))
				http.HandleFunc("/example.csv.sha256", staticServeHandler("example.csv.sha256"))
				http.HandleFunc("/example_refresh.csv", staticServeHandler("example_refresh.csv"))
				http.HandleFunc("/example_refresh.csv.sha256", staticServeHandler("example_refresh.csv.sha256"))
				log.Fatal(http.ListenAndServe(":10000", nil))
			}()
			// wait for initialization and listening to requests
			for {
				conn, _ := net.DialTimeout("tcp", net.JoinHostPort("", "10000"), time.Millisecond*time.Duration(10))
				if conn != nil {
					conn.Close()
					break
				}
			}
			if err := tt.s.refreshMapping(); (err != nil) != tt.wantErr {
				t.Errorf("server.refreshMapping_initial() error = %v, wantErr %v", err, tt.wantErr)
			}

			tt.s.param.proxy.Mapping.url = tt.newMap

			if err := tt.s.refreshMapping(); (err != nil) != tt.wantErr {
				t.Errorf("server.refreshMapping_reload() error = %v, wantErr %v", err, tt.wantErr)
			}
			got := 0
			for k := range tt.s.mapping {
				for range tt.s.mapping[k] {
					got = got + 1
				}
			}
			if got != tt.want {
				t.Errorf("server.refreshMapping_reload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func staticServeHandlerRefresh(filename string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, _ := os.Open("./examples/" + filename)
		defer file.Close()

		io.Copy(w, file)

		return
	}
}
