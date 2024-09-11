package jxpath

import "testing"

func TestFormatInteger(t *testing.T) {
	type args struct {
		x      float64
		format string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			args: args{
				x:      12345,
				format: "####0",
			},
			want: "12345",
		},
		{
			name: "test2",
			args: args{
				x:      12345,
				format: "###,#0",
			},
			want: "1,23,45",
		},
		{
			name: "test3",
			args: args{
				x:      12345,
				format: "##,#,#0",
			},
			want: "12,3,45",
		},
		{
			name: "test4",
			args: args{
				x:      12345,
				format: ",##,#,#0",
			},
			want: ",12,3,45",
		},
		{
			name: "test5",
			args: args{
				x:      12345,
				format: "w",
			},
			want: "twelve thousand, three hundred and forty-five",
		},
		{
			name: "test6",
			args: args{
				x:      12345,
				format: "W",
			},
			want: "TWELVE THOUSAND, THREE HUNDRED AND FORTY-FIVE",
		},
		{
			name: "test7",
			args: args{
				x:      12345,
				format: "Ww",
			},
			want: "Twelve Thousand, Three Hundred and Forty-Five",
		},
		{
			name: "test8",
			args: args{
				x:      12345,
				format: "w;o",
			},
			want: "twelve thousand, three hundred and forty-fifth",
		},
		{
			name: "test9",
			args: args{
				x:      1,
				format: "w;o",
			},
			want: "first",
		},
		{
			name: "test10",
			args: args{
				x:      10,
				format: "w;o",
			},
			want: "tenth",
		},
		{
			name: "test11",
			args: args{
				x:      11,
				format: "w;o",
			},
			want: "eleventh",
		},
		{
			name: "test12",
			args: args{
				x:      20,
				format: "w;o",
			},
			want: "twentieth",
		},
		{
			name: "test13",
			args: args{
				x:      21,
				format: "w;o",
			},
			want: "twenty-first",
		},
		{
			name: "test14",
			args: args{
				x:      91,
				format: "w;o",
			},
			want: "ninety-first",
		},
		{
			name: "test15",
			args: args{
				x:      91,
				format: "w;o",
			},
			want: "ninety-first",
		},
		{
			name: "test16",
			args: args{
				x:      100,
				format: "w;o",
			},
			want: "one hundredth",
		},
		{
			name: "test17",
			args: args{
				x:      101,
				format: "w;o",
			},
			want: "one hundred and first",
		},
		{
			name: "test18",
			args: args{
				x:      111,
				format: "w;o",
			},
			want: "one hundred and eleventh",
		},
		{
			name: "test19",
			args: args{
				x:      1000,
				format: "w;o",
			},
			want: "one thousandth",
		},
		{
			name: "test20",
			args: args{
				x:      10000,
				format: "w;o",
			},
			want: "ten thousandth",
		},
		{
			name: "test21",
			args: args{
				x:      1000000,
				format: "w;o",
			},
			want: "one millionth",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
