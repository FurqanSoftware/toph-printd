package main

import (
	_ "embed"
	"fmt"
	"log"
	"strings"

	"github.com/fatih/color"
	"github.com/signintech/gopdf"
)

type PDFBuilder struct {
	cfg Config
}

func (b PDFBuilder) Build(name string, pr Print) error {
	pdf := gopdf.GoPdf{}

	pagesize := gopdfPageSizes[b.cfg.Printer.PageSize]
	pdf.Start(gopdf.Config{PageSize: *pagesize})

	pdf.SetMargins(b.cfg.Printd.MarginLeft, b.cfg.Printd.MarginTop, b.cfg.Printd.MarginRight, b.cfg.Printd.MarginBottom)

	pdf.AddPage()
	err := pdf.AddTTFFontData("Ubuntu Mono", ubuntuMonoR)
	if err != nil {
		log.Fatalln(color.RedString("[E]"), err.Error())
		return nil
	}
	err = pdf.SetFont("Ubuntu Mono", "", b.cfg.Printd.FontSize)
	if err != nil {
		return err
	}

	linesperpage := int((pagesize.H - (b.cfg.Printd.MarginTop + b.cfg.Printd.MarginBottom)) / b.cfg.Printd.LineHeight)

	header := pr.Header
	headerextra := strings.TrimSpace(b.cfg.Printd.HeaderExtra)
	if headerextra != "" {
		header += " · " + headerextra
	}
	headerlines, err := pdf.SplitText(header, pagesize.W-pdf.MarginLeft()-pdf.MarginRight())
	if err != nil {
		return err
	}

	linesperpage -= len(headerlines) + 2

	lines, err := pdf.SplitText(b.tabToSpaces(pr.Content), pagesize.W-pdf.MarginLeft()-pdf.MarginRight())
	if err != nil {
		return err
	}

	if b.cfg.Printd.ReduceBlankLines {
		lines = reduceBlankLines(lines)
	}

	npages := int((len(lines) + (linesperpage)) / linesperpage)

	pageno := 0
	for i, l := range lines {
		y := pdf.GetY()
		if i > 0 {
			b.newLine(&pdf)
		}
		if i == 0 || pdf.GetY() < y {
			// New page
			pageno++
			b.header(&pdf, headerlines, i == 0, pageno, npages)
		}
		err = pdf.Cell(nil, l)
		if err != nil {
			return err
		}
	}

	return pdf.WritePdf(name)
}

func (b PDFBuilder) header(pdf *gopdf.GoPdf, lines []string, first bool, pageno, npages int) error {
	for _, l := range lines {
		err := pdf.Cell(nil, l)
		if err != nil {
			return err
		}
		b.newLine(pdf)
	}
	pdf.SetTextColor(92, 104, 115)
	var err error
	if first {
		err = pdf.Cell(nil, fmt.Sprintf("%d/%d · %s", pageno, npages, b.cfg.Toph.BaseURL))
	} else {
		err = pdf.Cell(nil, fmt.Sprintf("%d/%d", pageno, npages))
	}
	if err != nil {
		return err
	}
	pdf.SetTextColor(0, 0, 0)
	b.newLine(pdf)
	b.newLine(pdf)
	return nil
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
