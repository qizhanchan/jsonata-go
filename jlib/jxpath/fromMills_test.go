package jxpath

import (
	"strconv"
	"testing"
)

func TestFromMills(t *testing.T) {
	type args struct {
		unixMills int
		format    string
		timezone  string
	}
	tests := []struct {
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases. $fromMillis(1204405500000, '[Y]-[M01]-[D01]T[H01]:[m]:[s].[f001][Z0101t]', '+0530')
		{
			args: args{
				unixMills: 1204405500000,
				format:    "[Y]-[M01]-[D01]T[H01]:[m]:[s].[f001][Z0101t]",
				timezone:  "+0530",
			},
			want:    "2008-03-02T02:35:00.000+0530",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		testName := strconv.Itoa(tt.args.unixMills) + tt.args.format + tt.args.timezone
		t.Run(testName, func(t *testing.T) {
			got, err := FromMills(tt.args.unixMills, tt.args.format, tt.args.timezone)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromMills() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FromMills() got = %v, want %v", got, tt.want)
			}
		})
	}
}
