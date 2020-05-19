package main

import (
	"reflect"
	"testing"
)

func TestSeq(t *testing.T) {
	tests := []struct {
		min  int
		max  int
		want []int
	}{
		{0, 0, []int{0}},
		{0, 5, []int{0, 1, 2, 3, 4, 5}},
		{-3, 1, []int{-3, -2, -1, 0, 1}},
		{3, -2, []int{3, 2, 1, 0, -1, -2}},
	}

	for _, test := range tests {
		s := seq(test.min, test.max)
		if !reflect.DeepEqual(s, test.want) {
			t.Errorf("wrong sequence, got: %v, want: %v", s, test.want)
		}
	}
}
