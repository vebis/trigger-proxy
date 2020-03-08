package main

import (
	"bytes"
	"io"
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
		name       string
		args       args
		wantStdout string
		wantErr    bool
	}{
		{"single_repo", args{file: strings.NewReader("git://reposerver/repo;branch;job;")}, "Successfully read mappings: 1\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			if err := ParseMappingFile(tt.args.file, stdout); (err != nil) != tt.wantErr {
				t.Errorf("ParseMappingFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotStdout := stdout.String(); gotStdout != tt.wantStdout {
				t.Errorf("ParseMappingFile() = %v, want %v", gotStdout, tt.wantStdout)
			}
		})
	}
}
