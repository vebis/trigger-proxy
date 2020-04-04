package proxy

import (
	"io"
	"reflect"
	"strings"
	"testing"
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
						MappingFile: "./examples/example.csv",
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
						MappingFile:  "./examples/example_fm.csv",
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
