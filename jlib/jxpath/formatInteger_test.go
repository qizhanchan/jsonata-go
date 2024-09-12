package jxpath

import (
	"strconv"
	"testing"
)

func TestFormatInteger(t *testing.T) {
	type args struct {
		x      float64
		format string
	}
	tests := []struct {
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{

			args: args{
				x:      12345,
				format: "####0",
			},
			want: "12345",
		},
		{

			args: args{
				x:      12345,
				format: "###,#0",
			},
			want: "1,23,45",
		},
		{

			args: args{
				x:      12345,
				format: "##,#,#0",
			},
			want: "12,3,45",
		},
		{

			args: args{
				x:      12345,
				format: ",##,#,#0",
			},
			want: ",12,3,45",
		},
		{

			args: args{
				x:      12345,
				format: "w",
			},
			want: "twelve thousand, three hundred and forty-five",
		},
		{

			args: args{
				x:      12345,
				format: "W",
			},
			want: "TWELVE THOUSAND, THREE HUNDRED AND FORTY-FIVE",
		},
		{

			args: args{
				x:      12345,
				format: "Ww",
			},
			want: "Twelve Thousand, Three Hundred and Forty-Five",
		},
		{

			args: args{
				x:      12345,
				format: "w;o",
			},
			want: "twelve thousand, three hundred and forty-fifth",
		},
		{

			args: args{
				x:      1,
				format: "w;o",
			},
			want: "first",
		},
		{

			args: args{
				x:      10,
				format: "w;o",
			},
			want: "tenth",
		},
		{

			args: args{
				x:      11,
				format: "w;o",
			},
			want: "eleventh",
		},
		{

			args: args{
				x:      20,
				format: "w;o",
			},
			want: "twentieth",
		},
		{

			args: args{
				x:      21,
				format: "w;o",
			},
			want: "twenty-first",
		},
		{

			args: args{
				x:      91,
				format: "w;o",
			},
			want: "ninety-first",
		},
		{

			args: args{
				x:      91,
				format: "w;o",
			},
			want: "ninety-first",
		},
		{

			args: args{
				x:      100,
				format: "w;o",
			},
			want: "one hundredth",
		},
		{

			args: args{
				x:      101,
				format: "w;o",
			},
			want: "one hundred and first",
		},
		{

			args: args{
				x:      111,
				format: "w;o",
			},
			want: "one hundred and eleventh",
		},
		{

			args: args{
				x:      1000,
				format: "w;o",
			},
			want: "one thousandth",
		},
		{

			args: args{
				x:      10000,
				format: "w;o",
			},
			want: "ten thousandth",
		},
		{

			args: args{
				x:      1000000,
				format: "w;o",
			},
			want: "one millionth",
		},
		{

			args: args{
				x:      1,
				format: "i",
			},
			want: "i",
		},
		{

			args: args{
				x:      1,
				format: "I",
			},
			want: "I",
		},
		{

			args: args{
				x:      1,
				format: "I;o",
			},
			want:    "",
			wantErr: true,
		},
		{

			args: args{
				x:      1,
				format: "a",
			},
			want: "a",
		},
		{

			args: args{
				x:      1,
				format: "A",
			},
			want: "A",
		},
		{

			args: args{
				x:      1234,
				format: "a",
			},
			want: "aul",
		},
		{

			args: args{
				x:      1234,
				format: "au", // u 是无效的
			},
			want:    "",
			wantErr: true,
		},
		{

			args: args{
				x:      1234,
				format: "#０",
			},
			want: "１２３４",
		},
		{

			args: args{
				x:      1234,
				format: "#０0", // 不能同时包含全角和半角
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		testName := strconv.FormatFloat(tt.args.x, 'f', 1, 64) + "-" + tt.args.format
		t.Run(testName, func(t *testing.T) {
			got, err := FormatInteger(tt.args.x, tt.args.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("FormatInteger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FormatInteger() got = %v, want %v", got, tt.want)
			}
		})
	}
}
