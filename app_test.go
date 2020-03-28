package main

import (
	"io"
	"net/http"
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
		file      io.Reader
		filematch bool
	}
	tests := []struct {
		name    string
		args    args
		want    triggerMapping
		wantErr bool
	}{
		{
			"single_repo",
			args{file: strings.NewReader("git://repo/repo;branch;job;"), filematch: false},
			triggerMapping{map[string][]string{
				"git://repo/repo|branch": {"job"},
			}},
			false,
		},
		{
			"single_repo_filematch",
			args{file: strings.NewReader("git://repo/repo;branch;job;repo"), filematch: true},
			triggerMapping{map[string][]string{
				"git://repo/repo|branch|repo": {"job"},
			}},
			false,
		},
		{
			"single_repo_filematch_fail",
			args{file: strings.NewReader("git://repo/repo;branch;job"), filematch: true},
			triggerMapping{mapping: nil},
			true,
		},
		{
			"three_repos",
			args{file: strings.NewReader("git://repo/repo;branch;job\ngit://repo/repo;branch;job2\ngit://repo/repo2;branch;job"), filematch: false},
			triggerMapping{map[string][]string{
				"git://repo/repo|branch":  {"job", "job2"},
				"git://repo/repo2|branch": {"job"},
			}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMappingFile(tt.args.file, tt.args.filematch)
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

func TestParseGetRequest(t *testing.T) {
	reqSb, err := http.NewRequest("GET", "/?repo=git://repo&branch=devel", nil)
	if err != nil {
		t.Fatal(err)
	}
	reqS, err := http.NewRequest("GET", "/?repo=git://repo", nil)
	if err != nil {
		t.Fatal(err)
	}
	reqF, err := http.NewRequest("GET", "/?repo=git://repo&branch=master&file=1&file=a/b/2", nil)
	if err != nil {
		t.Fatal(err)
	}
	reqNr, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		r         *http.Request
		filematch bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		want2   []string
		wantErr bool
	}{
		{
			"common request with branch",
			args{r: reqSb, filematch: false},
			"git://repo",
			"devel",
			[]string{},
			false,
		},
		{
			"common request without branch",
			args{r: reqS, filematch: false},
			"git://repo",
			"master",
			[]string{},
			false,
		},
		{
			"common request with files",
			args{r: reqF, filematch: true},
			"git://repo",
			"master",
			[]string{"1", "a/b/2"},
			false,
		},
		{
			"common request with files without filematch",
			args{r: reqF, filematch: false},
			"git://repo",
			"master",
			[]string{},
			false,
		},
		{
			"no repo specified",
			args{r: reqNr, filematch: false},
			"",
			"",
			[]string{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := ParseGetRequest(tt.args.r, tt.args.filematch)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseGetRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseGetRequest() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ParseGetRequest() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("ParseGetRequest() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func Test_evalMappingKeys(t *testing.T) {
	type args struct {
		repo      string
		branch    string
		files     []string
		filematch bool
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			"simple_without_files",
			args{
				repo:      "git://repo/test",
				branch:    "master",
				files:     []string{},
				filematch: false,
			},
			[]string{"git://repo/test|master"},
			false,
		},
		{
			"simple_with_files",
			args{
				repo:      "git://repo/test",
				branch:    "master",
				files:     []string{"a", "b"},
				filematch: true,
			},
			[]string{"git://repo/test|master|a", "git://repo/test|master|b"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evalMappingKeys(tt.args.repo, tt.args.branch, tt.args.files, tt.args.filematch)
			if (err != nil) != tt.wantErr {
				t.Errorf("evalMappingKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("evalMappingKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_matchMappingKeysNoFileMatch(t *testing.T) {
	err := AssignMapping(strings.NewReader("git://repo/repo;branch;job;"), false)
	if err != nil {
		return
	}
	type args struct {
		keys      []string
		filematch bool
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			"simple",
			args{keys: []string{"git://repo/repo|branch"}, filematch: false},
			[]string{"job"},
			false,
		},
		{
			"no match",
			args{keys: []string{"git://repo/repo2|branch"}, filematch: false},
			[]string{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := matchMappingKeys(tt.args.keys, tt.args.filematch)
			if (err != nil) != tt.wantErr {
				t.Errorf("matchMappingKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("matchMappingKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}
