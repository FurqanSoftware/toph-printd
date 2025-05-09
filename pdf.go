package main

import (
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/FurqanSoftware/pog"
	"github.com/signintech/gopdf"
)

type PDFBuilder struct {
	cfg Config
}

type PDF struct {
	Name        string
	PageCount   int
	PageSkipped int
}

func (b PDFBuilder) Build(name string, pr Print) (PDF, error) {
	pdf := gopdf.GoPdf{}

	pagesize := gopdfPageSizes[b.cfg.Printer.PageSize]
	pdf.Start(gopdf.Config{PageSize: *pagesize})

	pdf.SetMargins(b.cfg.Printd.MarginLeft, b.cfg.Printd.MarginTop, b.cfg.Printd.MarginRight, b.cfg.Printd.MarginBottom)

	pdf.AddPage()
	err := pdf.AddTTFFontData("Ubuntu Mono", ubuntuMonoR)
	if err != nil {
		pog.Fatal(err)
		return PDF{}, nil
	}
	err = pdf.SetFont("Ubuntu Mono", "", b.cfg.Printd.FontSize)
	if err != nil {
		return PDF{}, err
	}

	linesperpage := int((pagesize.H - (b.cfg.Printd.MarginTop + b.cfg.Printd.MarginBottom)) / b.cfg.Printd.LineHeight)

	header := pr.Header
	header += " · " + pr.CreatedAt.In(time.Local).Format(time.DateTime)
	headerextra := strings.TrimSpace(b.cfg.Printd.HeaderExtra)
	if headerextra != "" {
		header += " · " + headerextra
	}
	var headerlines []string
	if header != "" {
		headerlines, err = pdf.SplitText(header, pagesize.W-pdf.MarginLeft()-pdf.MarginRight())
		if err != nil {
			return PDF{}, err
		}
	}

	linesperpage -= len(headerlines) + 2

	content := b.tabToSpaces(pr.Content)
	var lines []string
	if content != "" {
		lines, err = pdf.SplitText(content, pagesize.W-pdf.MarginLeft()-pdf.MarginRight())
		if err != nil {
			return PDF{}, err
		}
	}
	if len(lines) == 0 {
		lines = append(lines, "")
	}

	if b.cfg.Printd.ReduceBlankLines {
		lines = reduceBlankLines(lines)
	}

	npages := int((len(lines) + (linesperpage)) / linesperpage)

	pageno := 0
	for i, l := range lines {
		var newpage, atlimit, overlimit bool
		if pr.PageLimit != -1 {
			atlimit = pageno+1 >= pr.PageLimit
			overlimit = pageno+1 > pr.PageLimit
		}
		if i == 0 || b.isNextLineNewPage(&pdf, pagesize) {
			if overlimit {
				break
			}
			newpage = true
			pageno++
		}
		if i > 0 {
			b.newLine(&pdf)
		}
		if newpage {
			b.header(&pdf, headerlines, i == 0, pageno, npages, atlimit)
		}
		err = pdf.Cell(nil, l)
		if err != nil {
			return PDF{}, err
		}
	}

	if pageno > 0 {
		err = pdf.WritePdf(name)
		if err != nil {
			return PDF{}, err
		}
	}

	pageskipped := 0
	if pageno < npages {
		pageskipped = npages - pageno
	}

	return PDF{
		Name:        name,
		PageCount:   pageno,
		PageSkipped: pageskipped,
	}, err
}

func (b PDFBuilder) header(pdf *gopdf.GoPdf, lines []string, firstpage bool, pageno, npages int, atlimit bool) error {
	for _, l := range lines {
		err := pdf.Cell(nil, l)
		if err != nil {
			return err
		}
		b.newLine(pdf)
	}
	parts := []string{}
	parts = append(parts, fmt.Sprintf("%d/%d", pageno, npages))
	if atlimit {
		parts = append(parts, "Limit Reached")
	}
	if firstpage {
		parts = append(parts, "·", b.cfg.Toph.BaseURL)
	}
	pdf.SetTextColor(92, 104, 115)
	pdf.Cell(nil, strings.Join(parts, " "))
	pdf.SetTextColor(0, 0, 0)
	b.newLine(pdf)
	b.newLine(pdf)
	return nil
}

func (b PDFBuilder) isNextLineNewPage(pdf *gopdf.GoPdf, pagesize *gopdf.Rect) bool {
	return pdf.GetY()+float64(b.cfg.Printd.LineHeight) > pagesize.H-pdf.MarginBottom()
}

func (b PDFBuilder) newLine(pdf *gopdf.GoPdf) {
	pdf.SetNewXY(pdf.GetY()+float64(b.cfg.Printd.LineHeight), pdf.MarginLeft(), float64(b.cfg.Printd.LineHeight))
}

func (b PDFBuilder) tabToSpaces(t string) string {
	return strings.ReplaceAll(t, "\t", strings.Repeat(" ", b.cfg.Printd.TabSize))
}

var (
	gopdfPageSizes = map[PageSize]*gopdf.Rect{
		PageA4:     gopdf.PageSizeA4,
		PageLetter: gopdf.PageSizeLetter,
		PageLegal:  gopdf.PageSizeLegal,
	}
)

//go:embed UbuntuMono-R.ttf
var ubuntuMonoR []byte

func reduceBlankLines(lines []string) []string {
	reduced := lines[:0]
	lastblank := false
	for _, l := range lines {
		blank := strings.TrimSpace(l) == ""
		if lastblank && blank {
			continue
		}
		reduced = append(reduced, l)
		lastblank = blank
	}
	return reduced
}
