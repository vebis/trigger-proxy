package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func Test_server_handlePlainGet(t *testing.T) {
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name string
		s    server
		args args
		want int
	}{
		{
			"simple_match",
			server{
				mapping:    map[string][]string{"git://repo/magic/repo|branch|repo/file": {"job"}},
				timeKeeper: make(map[string]*time.Timer),
				param: parameters{
					proxy: proxy{
						QuietPeriod:  5,
						FileMatching: true,
						SemanticRepo: "git://repo/magic/",
					},
				},
			},
			args{w: httptest.NewRecorder(), r: httptest.NewRequest("GET", "/?repo=git://repo/magic/repo&branch=branch&file=file", nil)},
			http.StatusOK,
		},
		{
			"simple_nomatch",
			server{
				mapping:    map[string][]string{"git://repo/magic/repo|branch|repo/file": {"job"}},
				timeKeeper: make(map[string]*time.Timer),
				param: parameters{
					proxy: proxy{
						QuietPeriod:  5,
						FileMatching: true,
						SemanticRepo: "git://repo/magic/",
					},
				},
			},
			args{w: httptest.NewRecorder(), r: httptest.NewRequest("GET", "/?repo=git://repo/magic/repa&branch=branch&file=file", nil)},
			http.StatusNotFound,
		},
		{
			"bad_request",
			server{
				mapping:    map[string][]string{"git://repo/magic/repo|branch|repo/file": {"job"}},
				timeKeeper: make(map[string]*time.Timer),
				param: parameters{
					proxy: proxy{
						QuietPeriod:  5,
						FileMatching: true,
						SemanticRepo: "git://repo/magic/",
					},
				},
			},
			args{w: httptest.NewRecorder(), r: httptest.NewRequest("GET", "/?branch=branch&file=file", nil)},
			http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(tt.s.handlePlainGet())
			handler.ServeHTTP(tt.args.w, tt.args.r)
			if status := tt.args.w.Result().StatusCode; status != tt.want {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}
		})
	}
}

func Test_server_handleJSONPost(t *testing.T) {
	body := strings.NewReader(`{
		"object_kind": "push",
		"before": "95790bf891e76fee5e1747ab589903a6a1f80f22",
		"after": "da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
		"ref": "refs/heads/branch",
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
		  "git_ssh_url":"git@repo:magic/repo.git",
		  "git_http_url":"http://repo/magic/repo.git",
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
			"added": ["file"],
			"modified": [],
			"removed": []
		  }
		]
	  }`)
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name     string
		s        server
		args     args
		wantHTTP int
		wantHits int
	}{
		{
			"semantic_match",
			server{
				mapping:    map[string][]string{"git@repo:magic/repo.git|branch|repo/file": {"job"}},
				timeKeeper: make(map[string]*time.Timer),
				param: parameters{
					proxy: proxy{
						QuietPeriod:  5,
						FileMatching: true,
						SemanticRepo: "git@repo:magic/",
					},
				},
			},
			args{w: httptest.NewRecorder(), r: httptest.NewRequest("POST", "/json", strings.NewReader(`{`))},
			http.StatusBadRequest,
			0,
		},
		{
			"bad_request",
			server{
				mapping:    map[string][]string{"git@repo:magic/repo.git|branch|repo/file": {"job"}},
				timeKeeper: make(map[string]*time.Timer),
				param: parameters{
					proxy: proxy{
						QuietPeriod:  5,
						FileMatching: true,
						SemanticRepo: "git@repo:magic/",
					},
				},
			},
			args{w: httptest.NewRecorder(), r: httptest.NewRequest("POST", "/json", body)},
			http.StatusOK,
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(tt.s.handleJSONPost())
			handler.ServeHTTP(tt.args.w, tt.args.r)
			if status := tt.args.w.Result().StatusCode; status != tt.wantHTTP {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}
			hits := len(tt.s.timeKeeper)
			if tt.wantHits != hits {
				t.Errorf("handler returned wrong status code: got %v want %v", tt.wantHits, hits)
			}
		})
	}
}
