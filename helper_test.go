package proxy

import (
	"reflect"
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
			if got := buildMappingKey(tt.args.keys); got != tt.want {
				t.Errorf("BuildMappingKey() = %v, want %v", got, tt.want)
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
				t.Errorf("CreateJobURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeLastRune(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"simple",
			args{s: "ab"},
			"a",
		},
		{
			"simple_oneleter",
			args{s: "a"},
			"",
		},
		{
			"simple_zeroletter",
			args{s: ""},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeLastRune(tt.args.s); got != tt.want {
				t.Errorf("removeLastRune() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_evalMappingKeys(t *testing.T) {
	type args struct {
		repo         string
		branch       string
		files        []string
		filematch    bool
		semanticrepo string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"simple_without_files",
			args{
				repo:         "git://repo/test",
				branch:       "master",
				files:        []string{},
				filematch:    false,
				semanticrepo: "",
			},
			[]string{"git://repo/test|master"},
		},
		{
			"simple_with_files",
			args{
				repo:         "git://repo/test",
				branch:       "master",
				files:        []string{"a", "b"},
				filematch:    true,
				semanticrepo: "",
			},
			[]string{"git://repo/test|master|a", "git://repo/test|master|b"},
		},
		{
			"matching_semantic_repo",
			args{
				repo:         "git://repo/magic/test.git",
				branch:       "master",
				files:        []string{"a", "b"},
				filematch:    true,
				semanticrepo: "git://repo/magic/",
			},
			[]string{"git://repo/magic/test.git|master|test/a", "git://repo/magic/test.git|master|test/b"},
		},
		{
			"non_matching_semantic_repo",
			args{
				repo:         "git://repo/test.git",
				branch:       "master",
				files:        []string{"a", "b"},
				filematch:    true,
				semanticrepo: "git://repo/magic/",
			},
			[]string{"git://repo/test.git|master|a", "git://repo/test.git|master|b"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evalMappingKeys(tt.args.repo, tt.args.branch, tt.args.files, tt.args.filematch, tt.args.semanticrepo)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("evalMappingKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}
