package main

import (
	"reflect"
	"sort"
	"testing"
	"time"
)

func Test_matchMappingKeysNoFileMatch(t *testing.T) {
	type args struct {
		keys      []string
		filematch bool
	}
	tests := []struct {
		name    string
		server  server
		args    args
		want    []string
		wantErr bool
	}{
		{
			"simple",
			server{
				mapping:    map[string][]string{"git://repo/repo|branch": {"job"}},
				timeKeeper: make(map[string]*time.Timer),
			},
			args{keys: []string{"git://repo/repo|branch"}, filematch: false},
			[]string{"job"},
			false,
		},
		{
			"no match",
			server{
				mapping:    map[string][]string{"git://repo/repo|branch": {"job"}},
				timeKeeper: make(map[string]*time.Timer),
			},
			args{keys: []string{"git://repo/repo2|branch"}, filematch: false},
			[]string{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.server.matchMappingKeys(tt.args.keys, tt.args.filematch)
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

func Test_matchMappingKeysFileMatch(t *testing.T) {
	type args struct {
		keys      []string
		filematch bool
	}
	tests := []struct {
		name    string
		server  server
		args    args
		want    []string
		wantErr bool
	}{
		{
			"simple_direct_hit",
			server{
				mapping:    map[string][]string{"git://repo/repo|branch|cli": {"job"}},
				timeKeeper: make(map[string]*time.Timer),
			},
			args{keys: []string{"git://repo/repo|branch|cli"}, filematch: true},
			[]string{"job"},
			false,
		},
		{
			"simple_indirect_hit",
			server{
				mapping:    map[string][]string{"git://repo/repo|branch|cli": {"job"}},
				timeKeeper: make(map[string]*time.Timer),
			},
			args{keys: []string{"git://repo/repo|branch|cli/other"}, filematch: true},
			[]string{"job"},
			false,
		},
		{
			"no match",
			server{
				mapping:    make(map[string][]string),
				timeKeeper: make(map[string]*time.Timer),
			},
			args{keys: []string{"git://repo/repo2|branch|bla"}, filematch: true},
			[]string{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.server.matchMappingKeys(tt.args.keys, tt.args.filematch)
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

func Test_server_processMatching(t *testing.T) {
	type args struct {
		repo   string
		branch string
		files  []string
	}
	tests := []struct {
		name       string
		s          server
		args       args
		wantTimers []string
		wantErr    bool
	}{
		{
			"simple_match",
			server{
				mapping:    map[string][]string{"git://repo/repo|branch": {"job", "job2"}},
				timeKeeper: make(map[string]*time.Timer),
				param: parameters{
					proxy: proxy{
						QuietPeriod:  5,
						FileMatching: false,
						SemanticRepo: "",
					},
				},
			},
			args{repo: "git://repo/repo", branch: "branch", files: []string{}},
			[]string{"job", "job2"},
			false,
		},
		{
			"simple_https_match",
			server{
				mapping:    map[string][]string{"https://repo/repo|branch": {"job", "job2"}},
				timeKeeper: make(map[string]*time.Timer),
				param: parameters{
					proxy: proxy{
						QuietPeriod:  5,
						FileMatching: false,
						SemanticRepo: "",
					},
				},
			},
			args{repo: "https://repo/repo", branch: "branch", files: []string{}},
			[]string{"job", "job2"},
			false,
		},
		{
			"simple_ssh_match",
			server{
				mapping:    map[string][]string{"git@repo:repo|branch": {"job", "job2"}},
				timeKeeper: make(map[string]*time.Timer),
				param: parameters{
					proxy: proxy{
						QuietPeriod:  5,
						FileMatching: false,
						SemanticRepo: "",
					},
				},
			},
			args{repo: "git@repo:repo", branch: "branch", files: []string{}},
			[]string{"job", "job2"},
			false,
		},
		{
			"semantic_ssh_match",
			server{
				mapping:    map[string][]string{"git@repo:magic/repo|branch|repo/file": {"job", "job2"}},
				timeKeeper: make(map[string]*time.Timer),
				param: parameters{
					proxy: proxy{
						QuietPeriod:  5,
						FileMatching: true,
						SemanticRepo: "git@repo:magic/",
					},
				},
			},
			args{repo: "git@repo:magic/repo", branch: "branch", files: []string{"file"}},
			[]string{"job", "job2"},
			false,
		},
		{
			"simple_nomatch",
			server{
				mapping:    map[string][]string{"git://repo/repo|branch": {"job"}},
				timeKeeper: make(map[string]*time.Timer),
				param: parameters{
					proxy: proxy{
						QuietPeriod:  5,
						FileMatching: false,
						SemanticRepo: "",
					},
				},
			},
			args{repo: "git://repo/repo", branch: "branch2", files: []string{}},
			[]string{},
			true,
		},
		{
			"filematch_exact_match",
			server{
				mapping:    map[string][]string{"git://repo/repo|branch|folder": {"job"}},
				timeKeeper: make(map[string]*time.Timer),
				param: parameters{
					proxy: proxy{
						QuietPeriod:  5,
						FileMatching: true,
						SemanticRepo: "",
					},
				},
			},
			args{repo: "git://repo/repo", branch: "branch", files: []string{"folder"}},
			[]string{"job"},
			false,
		},
		{
			"filematch_greedy_match",
			server{
				mapping:    map[string][]string{"git://repo/repo|branch|folder": {"job"}},
				timeKeeper: make(map[string]*time.Timer),
				param: parameters{
					proxy: proxy{
						QuietPeriod:  5,
						FileMatching: true,
						SemanticRepo: "",
					},
				},
			},
			args{repo: "git://repo/repo", branch: "branch", files: []string{"foldergreedy"}},
			[]string{"job"},
			false,
		},
		{
			"filematch_no_match",
			server{
				mapping:    map[string][]string{"git://repo/repo|branch|folder": {"job"}},
				timeKeeper: make(map[string]*time.Timer),
				param: parameters{
					proxy: proxy{
						QuietPeriod:  5,
						FileMatching: true,
						SemanticRepo: "",
					},
				},
			},
			args{repo: "git://repo/repo", branch: "branch", files: []string{"foldes"}},
			[]string{},
			true,
		},
		{
			"filematch_semanticrepo_exact_match",
			server{
				mapping: map[string][]string{
					"git://repo/repo|branch|folder":            {"job"},
					"git://repo/magic/repo|branch|repo/folder": {"job2"},
				},
				timeKeeper: make(map[string]*time.Timer),
				param: parameters{
					proxy: proxy{
						QuietPeriod:  5,
						FileMatching: true,
						SemanticRepo: "git://repo/magic/",
					},
				},
			},
			args{repo: "git://repo/magic/repo", branch: "branch", files: []string{"folder"}},
			[]string{"job2"},
			false,
		},
		{
			"filematch_semanticrepo_greedy_match",
			server{
				mapping: map[string][]string{
					"git://repo/repo|branch|folder":            {"job"},
					"git://repo/magic/repo|branch|repo/folder": {"job2"},
				},
				timeKeeper: make(map[string]*time.Timer),
				param: parameters{
					proxy: proxy{
						QuietPeriod:  5,
						FileMatching: true,
						SemanticRepo: "git://repo/magic/",
					},
				},
			},
			args{repo: "git://repo/magic/repo", branch: "branch", files: []string{"folder/test"}},
			[]string{"job2"},
			false,
		},
		{
			"filematch_non_semanticrepo_exact_match",
			server{
				mapping: map[string][]string{
					"git://repo/repo|branch|folder":            {"job"},
					"git://repo/magic/repo|branch|repo/folder": {"job2"},
				},
				timeKeeper: make(map[string]*time.Timer),
				param: parameters{
					proxy: proxy{
						QuietPeriod:  5,
						FileMatching: true,
						SemanticRepo: "git://repo/magic/",
					},
				},
			},
			args{repo: "git://repo/repo", branch: "branch", files: []string{"folder"}},
			[]string{"job"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.processMatching(tt.args.repo, tt.args.branch, tt.args.files); (err != nil) != tt.wantErr {
				t.Errorf("server.processMatching() error = %v, wantErr %v", err, tt.wantErr)
			}
			got := make([]string, 0, len(tt.s.timeKeeper))
			for k := range tt.s.timeKeeper {
				got = append(got, k)
			}
			sort.Strings(got)
			sort.Strings(tt.wantTimers)
			if !reflect.DeepEqual(got, tt.wantTimers) {
				t.Errorf("server.processMatching() = %v, want %v", got, tt.wantTimers)
			}
		})
	}
}
