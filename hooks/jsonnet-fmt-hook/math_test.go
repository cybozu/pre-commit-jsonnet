package main

import (
	"testing"
)

func TestMin(t *testing.T) {
	params := []struct {
		values [2]int
		want   int
	}{
		{
			values: [2]int{1, 2},
			want:   1,
		},
		{
			values: [2]int{1, -1},
			want:   -1,
		},
		{
			values: [2]int{2, 2},
			want:   2,
		},
	}

	for _, param := range params {
		got := min(param.values[0], param.values[1])
		if got != param.want {
			t.Errorf("args='%q', want=%v, got=%v", param.values, param.want, got)
		}
	}
}
