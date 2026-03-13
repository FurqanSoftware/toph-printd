package main

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
			assert.Equal(t, c.want, got)
		})
	}
}

func TestBuildPageLimit(t *testing.T) {
	cfg := Config{}
	cfg.initDefaults()
	b := PDFBuilder{cfg: cfg}

	longContent := strings.Repeat("Line of text\n", 500)

	t.Run("no limit", func(t *testing.T) {
		pdf, err := b.Build("test_no_limit.pdf", Print{
			Header:    "Test",
			Content:   longContent,
			PageLimit: -1,
		})
		assert.NoError(t, err)
		assert.Greater(t, pdf.PageCount, 1)
		assert.Equal(t, 0, pdf.PageSkipped)
		os.Remove("test_no_limit.pdf")
	})

	t.Run("limit 1", func(t *testing.T) {
		pdf, err := b.Build("test_limit_1.pdf", Print{
			Header:    "Test",
			Content:   longContent,
			PageLimit: 1,
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, pdf.PageCount)
		assert.Greater(t, pdf.PageSkipped, 0)
		os.Remove("test_limit_1.pdf")
	})

	t.Run("limit 2", func(t *testing.T) {
		pdf, err := b.Build("test_limit_2.pdf", Print{
			Header:    "Test",
			Content:   longContent,
			PageLimit: 2,
		})
		assert.NoError(t, err)
		assert.Equal(t, 2, pdf.PageCount)
		assert.Greater(t, pdf.PageSkipped, 0)
		os.Remove("test_limit_2.pdf")
	})

	t.Run("limit exceeds pages", func(t *testing.T) {
		pdf, err := b.Build("test_limit_high.pdf", Print{
			Header:    "Test",
			Content:   "Short content",
			PageLimit: 100,
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, pdf.PageCount)
		assert.Equal(t, 0, pdf.PageSkipped)
		os.Remove("test_limit_high.pdf")
	})
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
			assert.Equal(t, c.want, got)
		})
	}
}
