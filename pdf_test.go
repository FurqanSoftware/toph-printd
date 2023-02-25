package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReduceBlankLines(t *testing.T) {
	for _, c := range []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "no blank lines",
			in: []string{
				"lorem ipsum",
				"dolor",
				"sit amet",
			},
			want: []string{
				"lorem ipsum",
				"dolor",
				"sit amet",
			},
		},
		{
			name: "single blank lines with spaces",
			in: []string{
				"lorem ipsum",
				"		",
				"dolor",
				"  ",
				"sit",
				"",
				"amet",
			},
			want: []string{
				"lorem ipsum",
				"		",
				"dolor",
				"  ",
				"sit",
				"",
				"amet",
			},
		},
		{
			name: "consecutive blank lines with spaces",
			in: []string{
				"lorem ipsum",
				"		",
				"dolor",
				"  ",
				"",
				"",
				"sit",
				"",
				"   ",
				"amet",
			},
			want: []string{
				"lorem ipsum",
				"		",
				"dolor",
				"  ",
				"sit",
				"",
				"amet",
			},
		},
		{
			name: "leading blank lines",
			in: []string{
				"",
				"",
				"    ",
				"",
				"lorem ipsum",
				"		",
				"dolor",
				"  ",
				"",
				"",
				"sit",
			},
			want: []string{
				"",
				"lorem ipsum",
				"		",
				"dolor",
				"  ",
				"sit",
			},
		},
		{
			name: "trailing blank lines",
			in: []string{
				"lorem ipsum",
				"		",
				"dolor",
				"  ",
				"",
				"",
				"sit",
				"       	",
				"",
				"    ",
				"",
			},
			want: []string{
				"lorem ipsum",
				"		",
				"dolor",
				"  ",
				"sit",
				"       	",
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := reduceBlankLines(c.in)
			diff := cmp.Diff(got, c.want)
			if diff != "" {
				t.Fatalf("want: %q\ngot: %q\ndiff:\n%v", c.want, got, diff)
			}
		})
	}
}
