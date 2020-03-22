package main

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestBuildMappingKey(t *testing.T) {
	type args struct {
		keys []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"simple_ab", args{keys: []string{"a", "b"}}, "a|b"},
		{"simple_abc", args{keys: []string{"a", "b", "c"}}, "a|b|c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildMappingKey(tt.args.keys); got != tt.want {
				t.Errorf("BuildMappingKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseMappingFile(t *testing.T) {
	type args struct {
		file io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    triggerMapping
		wantErr bool
	}{
		{
			"single_repo",
			args{file: strings.NewReader("git://reposerver/repo;branch;job;")},
			triggerMapping{map[string][]string{
				"git://reposerver/repo|branch": {"job"},
			}},
			false,
		},
		{
			"three_repos",
			args{file: strings.NewReader("git://reposerver/repo;branch;job\ngit://reposerver/repo;branch;job2\ngit://reposerver/repo2;branch;job")},
			triggerMapping{map[string][]string{
				"git://reposerver/repo|branch":  {"job", "job2"},
				"git://reposerver/repo2|branch": {"job"},
			}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMappingFile(tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMappingFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseMappingFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createJobURL(t *testing.T) {
	type args struct {
		jenkinsURL string
		job        string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"jenkins_url", args{jenkinsURL: "http://jenkins:8080", job: "test"}, "http://jenkins:8080/job/test/build",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createJobURL(tt.args.jenkinsURL, tt.args.job); got != tt.want {
				t.Errorf("createJobURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
