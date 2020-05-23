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

func Test_server_process(t *testing.T) {
	tests := []struct {
		name    string
		s       server
		fm      bool
		want    int
		wantErr bool
	}{
		{
			"example_parser_nofilematch",
			server{
				mappingSource: mappingFile{
					path: "./examples/example.csv",
				},
			},
			false,
			8,
			false,
		},
		{
			"example_parser_filematch",
			server{
				mappingSource: mappingFile{
					path: "./examples/example_fm.csv",
				},
			},
			true,
			2,
			false,
		},
		{
			"example_parser_nofilematch",
			server{
				mappingSource: mappingFile{
					path: "./examples/example-notexistent.csv",
				},
			},
			false,
			0,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _, err := tt.s.mappingSource.process(tt.fm)
			if (err != nil) != tt.wantErr {
				t.Errorf("mappingSource.process() error = %v, wantErr %v", err, tt.wantErr)
			}
			got := 0
			for k := range m {
				for range m[k] {
					got = got + 1
				}
			}
			if got != tt.want {
				t.Errorf("mappingSource.process() = %v, want %v", got, tt.want)
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
				mappingSource: mappingFile{
					path: "./examples/example.csv",
				},
			},
			8,
			false,
		},
		{
			"example_parser_filematch",
			server{
				mappingSource: mappingFile{
					path: "./examples/example_fm.csv",
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
		newmh   mappingHandler
		want    int
		wantErr bool
	}{
		{
			"refresh_mapping",
			server{
				mappingSource: mappingFile{
					path: "./examples/example.csv",
				},
			},

			mappingFile{
				path: "./examples/example_refresh.csv",
			},
			7,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.refreshMapping(); (err != nil) != tt.wantErr {
				t.Errorf("server.refreshMapping() error = %v, wantErr %v", err, tt.wantErr)
			}

			tt.s.mappingSource = tt.newmh
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
		newMap  mappingHandler
		want    int
		wantErr bool
	}{
		{
			"refresh_mapping",
			server{
				mappingSource: mappingURL{
					path: "http://localhost:10000/example.csv",
				},
			},
			mappingURL{
				path: "http://localhost:10000/example_refresh.csv",
			},
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

			tt.s.mappingSource = tt.newMap

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
