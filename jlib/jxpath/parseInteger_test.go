package jxpath

import "testing"

func TestParseInteger(t *testing.T) {
	type args struct {
		num    string
		format string
	}
	tests := []struct {
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				num:    "twenty-one",
				format: "w",
			},
			want: 21,
		},
		{
			args: args{
				num:    "twelve thousand, four hundred and seventy-six",
				format: "w",
			},
			want: 12476,
		},
		{
			args: args{
				num:    "123rd",
				format: "000;o",
			},
			want: 123,
		},
		{
			args: args{
				num:    "１２３４０",
				format: "###０",
			},
			want: 12340,
		},
		{
			args: args{
				num:    "1,200",
				format: "#,##0",
			},
			want: 1200,
		},
		{
			args: args{
				num:    "",
				format: "I",
			},
			want: 0,
		},
		{
			args: args{
				num:    "MCMLXXXIV",
				format: "I",
			},
			want: 1984,
		},
		{
			args: args{
				num:    "xcix",
				format: "i",
			},
			want: 99,
		},
		{
			args: args{
				num:    "NINETY-NINE",
				format: "W",
			},
			want: 99,
		},
		{
			args: args{
				num:    "one thousand trillion",
				format: "w",
			},
			want: 1e+15,
		},
	}
	for _, tt := range tests {
		name := tt.args.num + "_" + tt.args.format
		t.Run(name, func(t *testing.T) {
			got, err := ParseInteger(tt.args.num, tt.args.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInteger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseInteger() got = %v, want %v", got, tt.want)
			}
		})
	}
}
