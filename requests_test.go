package proxy

import (
	"net/http"
	"reflect"
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
