package test

import "testing"

func Test_print1(t *testing.T) {
	type args struct {
		x int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "old cov",
			args: args{
				x: 1,
			},
		},
		{
			name: "old cov",
			args: args{
				x: 2,
			},
		},
		{
			name: "old cov",
			args: args{
				x: 4,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			print1(tt.args.x)
		})
	}
}

func Test_print(t *testing.T) {
	type args struct {
		x int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "old cov",
			args: args{
				x: 3,
			},
		},
		{
			name: "old cov",
			args: args{
				x: 5,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			print2(tt.args.x)
		})
	}
}
