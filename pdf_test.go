package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestExpandTabs(t *testing.T) {
	b := PDFBuilder{cfg: Config{}}
	b.cfg.Printd.TabSize = 4

	for _, c := range []struct {
		name string
		in   string
		want string
	}{
		{
			name: "no tabs",
			in:   "hello world",
			want: "hello world",
		},
		{
			name: "tab at start",
			in:   "\thello",
			want: "    hello",
		},
		{
			name: "tab after one char",
			in:   "a\tb",
			want: "a   b",
		},
		{
			name: "tab after two chars",
			in:   "ab\tc",
			want: "ab  c",
		},
		{
			name: "tab after three chars",
			in:   "abc\td",
			want: "abc d",
		},
		{
			name: "tab after four chars",
			in:   "abcd\te",
			want: "abcd    e",
		},
		{
			name: "multiple tabs",
			in:   "\t\thello",
			want: "        hello",
		},
		{
			name: "mixed content with tabs",
			in:   "if\t(x)\t{",
			want: "if  (x) {",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := b.expandTabs(c.in)
			if got != c.want {
				t.Fatalf("expandTabs(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

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
