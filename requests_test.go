package main

import (
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func Test_parseGetRequest(t *testing.T) {
	reqSb, err := http.NewRequest("GET", "/?repo=git://repo&branch=devel", nil)
	if err != nil {
		t.Fatal(err)
	}
	reqS, err := http.NewRequest("GET", "/?repo=git://repo", nil)
	if err != nil {
		t.Fatal(err)
	}
	reqF, err := http.NewRequest("GET", "/?repo=git://repo&branch=master&files=1&files=a/b/2", nil)
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
			got, got1, got2, err := parseGetRequest(tt.args.r, tt.args.filematch)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseGetRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseGetRequest() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("parseGetRequest() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("parseGetRequest() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func Test_parseJSONRequest(t *testing.T) {
	body := strings.NewReader(`{
		"object_kind": "push",
		"before": "95790bf891e76fee5e1747ab589903a6a1f80f22",
		"after": "da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
		"ref": "refs/heads/master",
		"checkout_sha": "da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
		"user_id": 4,
		"user_name": "John Smith",
		"user_username": "jsmith",
		"user_email": "john@example.com",
		"user_avatar": "https://s.gravatar.com/avatar/d4c74594d841139328695756648b6bd6?s=8://s.gravatar.com/avatar/d4c74594d841139328695756648b6bd6?s=80",
		"project_id": 15,
		"project":{
		  "id": 15,
		  "name":"Diaspora",
		  "description":"",
		  "web_url":"http://example.com/mike/diaspora",
		  "avatar_url":null,
		  "git_ssh_url":"git@example.com:mike/diaspora.git",
		  "git_http_url":"http://example.com/mike/diaspora.git",
		  "namespace":"Mike",
		  "visibility_level":0,
		  "path_with_namespace":"mike/diaspora",
		  "default_branch":"master",
		  "homepage":"http://example.com/mike/diaspora",
		  "url":"git@example.com:mike/diaspora.git",
		  "ssh_url":"git@example.com:mike/diaspora.git",
		  "http_url":"http://example.com/mike/diaspora.git"
		},
		"repository":{
		  "name": "Diaspora",
		  "url": "git@example.com:mike/diaspora.git",
		  "description": "",
		  "homepage": "http://example.com/mike/diaspora",
		  "git_http_url":"http://example.com/mike/diaspora.git",
		  "git_ssh_url":"git@example.com:mike/diaspora.git",
		  "visibility_level":0
		},
		"commits": [
		  {
			"id": "b6568db1bc1dcd7f8b4d5a946b0b91f9dacd7327",
			"message": "Update Catalan translation to e38cb41.\n\nSee https://gitlab.com/gitlab-org/gitlab for more information",
			"title": "Update Catalan translation to e38cb41.",
			"timestamp": "2011-12-12T14:27:31+02:00",
			"url": "http://example.com/mike/diaspora/commit/b6568db1bc1dcd7f8b4d5a946b0b91f9dacd7327",
			"author": {
			  "name": "Jordi Mallach",
			  "email": "jordi@softcatala.org"
			},
			"added": ["CHANGELOG"],
			"modified": ["app/controller/application.rb"],
			"removed": ["README.md"]
		  },
		  {
			"id": "da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
			"message": "fixed readme",
			"title": "fixed readme",
			"timestamp": "2012-01-03T23:36:29+02:00",
			"url": "http://example.com/mike/diaspora/commit/da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
			"author": {
			  "name": "GitLab dev user",
			  "email": "gitlabdev@dv6700.(none)"
			},
			"added": ["CHANGELOG"],
			"modified": ["app/controller/application.rb"],
			"removed": []
		  }
		],
		"total_commits_count": 4
	  }`)
	reqSb, _ := http.NewRequest("POST", "/", body)
	type args struct {
		r         *http.Request
		filematch bool
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		want1   string
		want2   []string
		wantErr bool
	}{
		{
			"simple",
			args{r: reqSb, filematch: true},
			[]string{"http://example.com/mike/diaspora.git", "git@example.com:mike/diaspora.git"},
			"master",
			[]string{"CHANGELOG", "README.md", "app/controller/application.rb"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := parseJSONRequest(tt.args.r, tt.args.filematch)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseJSONRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseJSONRequest() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("parseJSONRequest() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("parseJSONRequest() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
