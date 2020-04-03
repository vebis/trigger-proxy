package proxy

import (
	"testing"
	"time"
)

func Test_server_createTimer(t *testing.T) {
	type args struct {
		job string
	}
	tests := []struct {
		name string
		s    server
		args args
		want int
	}{
		{
			"simple",
			server{
				mapping:    map[string][]string{"git://repo/repo|branch": {"job"}},
				timeKeeper: map[string]*time.Timer{"job": time.AfterFunc(time.Second*time.Duration(1), func() {})},
			},
			args{job: "job"},
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.createTimer(tt.args.job)
			got := len(tt.s.timeKeeper)
			if got != tt.want {
				t.Errorf("server_createTimer() got = %v, want %v", got, tt.want)
			}
		})
	}
}
