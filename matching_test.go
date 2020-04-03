package proxy

import (
	"reflect"
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
