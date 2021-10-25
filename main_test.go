package main

import (
	"io/ioutil"
	"reflect"
	"testing"
)

func TestPrint(t *testing.T) {
	type args struct {
		x int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "1",
			args: args{
				x: 1,
			},
		},
		{
			name: "3",
			args: args{
				x: 3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Print(tt.args.x); got != tt.want {
				t.Errorf("Print() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMap(t *testing.T) {
	main1, err := ioutil.ReadFile("./test/main1.go")
	if err != nil {
		return
	}
	main2, err := ioutil.ReadFile("./test/main2.go")
	if err != nil {
		return
	}
	type args struct {
		str1 string
		str2 string
	}
	tests := []struct {
		name string
		args args
		want map[int]int
	}{
		{
			name: "change",
			args: args{
				str1: `a
b
c
d`,
				str2: `a
b
cc
d`,
			},
			want: map[int]int{
				1: 1,
				2: 2,
				4: 4,
			},
		},
		{
			name: "insert",
			args: args{
				str1: `a
b
d`,
				str2: `a
b
c
d`,
			},
			want: map[int]int{
				1: 1,
				2: 2,
				3: 4,
			},
		},
		{
			name: "delete",
			args: args{
				str1: `a
b
c
d`,
				str2: `a
b
d`,
			},
			want: map[int]int{
				1: 1,
				2: 2,
				4: 3,
			},
		},
		{ //如果旧字符串最后一行跟新的不一致，会导致最后一组diff结果不太对，insert连到了一起，不过小问题
			name: "insert and delete",
			args: args{
				str1: string(main1),
				str2: string(main2),
			},
			want: map[int]int{
				1:  1,
				2:  2,
				3:  3,
				4:  4,
				6:  6,
				7:  7,
				8:  8,
				11: 11,
				12: 12,
				13: 15,
				14: 16,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetMap(tt.args.str1, tt.args.str2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
