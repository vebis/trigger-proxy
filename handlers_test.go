package proxy

import (
	"net/http"
	"net/http/httptest"
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
					QuietPeriod:  5,
					FileMatching: true,
					SemanticRepo: "git://repo/magic/",
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
					QuietPeriod:  5,
					FileMatching: true,
					SemanticRepo: "git://repo/magic/",
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
					QuietPeriod:  5,
					FileMatching: true,
					SemanticRepo: "git://repo/magic/",
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
