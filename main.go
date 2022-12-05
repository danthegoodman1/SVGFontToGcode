package main

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type (
	FontSVG struct {
		Glyphs           []SVGGlyph
		MissingGlyphXAdv float32
	}

	SVGGlyph struct {
		Char rune
		// How much to advance on the X-axis after drawing the character
		XAdv float32
		D    []SVGDrawInstruction
	}

	SVGDrawInstruction struct {
		Instruction SVGInstruction
		X           float32
		Y           float32
	}

	SVGInstruction rune
)

var (
	// Pen up, move to location
	SVGInstruction_MoveAbsolute SVGInstruction = 'M'
	// Pen down, move to locaiton
	SVGInstruction_LineAbsolute SVGInstruction = 'L'
)

func main() {
	log.Println("starting SVG font to Gcode")

	// I know it's deprecated but I am lazy
	fileNames := make([]string, 0)
	filepath.Walk("fonts", func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			fmt.Println("found file", path)
			fileNames = append(fileNames, path)
			svg, err := svgToGcode(path)
			if err != nil {
				return fmt.Errorf("error in svgToGcode for %s: %w", path, err)
			}
			fmt.Printf("%+v\n", svg)
			// FIXME: REMOVE THIS, for development
			os.Exit(0)
		}
		return nil
	})

	log.Printf("parsing out %d files to out/", len(fileNames))
}

func svgToGcode(filePath string) (*FontSVG, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error in os.Readfile: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(fileContent))
	if err != nil {
		return nil, fmt.Errorf("error in NewDocumentFromReader: %w", err)
	}

	fontSVG := &FontSVG{
		Glyphs: []SVGGlyph{},
	}

	glyphs := doc.Find("glyph")
	glyphs.Each(func(i int, selection *goquery.Selection) {
		// it gets auto converted from HTML to real unicode, nice
		unicodeVal, exists := selection.Attr("unicode")
		xAdv, xadvExists := selection.Attr("horiz-adv-x")
		drawing, drawingExists := selection.Attr("d")
		if exists && xadvExists && drawingExists {
			x, err := strconv.ParseFloat(xAdv, 32)
			if err != nil {
				// This is lazy but I don't care right now
				panic(fmt.Errorf("error in strconv.ParseInt: %w", err))
			}
			xf := float32(x)

			// Parse out the drawing instructions
			fmt.Println("runes", []rune(unicodeVal))
			if runes := []rune(unicodeVal); len(runes) == 0 || runes[0] > 126 {
				// We don't want non-ascii values
				return
			}
			instructions, err := ParseSVGDrawingInstructions(drawing)
			if err != nil {
				panic(err)
			}

			fontSVG.Glyphs = append(fontSVG.Glyphs, SVGGlyph{
				Char: []rune(unicodeVal)[0],
				XAdv: xf,
				D:    instructions,
			})
		}
	})

	missingGlyph := doc.Find("missing-glyph")
	if xAdv, exists := missingGlyph.First().Attr("horiz-adv-x"); exists {
		x, err := strconv.ParseFloat(xAdv, 32)
		if err != nil {
			// This is lazy but I don't care right now
			panic(fmt.Errorf("error in strconv.ParseInt: %w", err))
		}
		fontSVG.MissingGlyphXAdv = float32(x)
	}

	return fontSVG, nil
}

func ParseSVGDrawingInstructions(raw string) ([]SVGDrawInstruction, error) {
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, " ")

	// Each part is always format `{instruction} {X} {Y}`, separated by spaces
	instructions := []SVGDrawInstruction{}
	workingPart := SVGDrawInstruction{}
	for i, part := range parts {
		fmt.Println(i, len(parts), part)
		switch i % 3 {
		case 0:
			// Instruction
			runes := []rune(part)
			if len(runes) == 0 {
				// We don't want this character
				return nil, nil
			}
			workingPart.Instruction = SVGInstruction([]rune(part)[0])
		case 1:
			// X
			x, err := strconv.ParseFloat(part, 32)
			if err != nil {
				return nil, fmt.Errorf("error in x strconv.ParseFloat for %s: %w", part, err)
			}
			workingPart.X = float32(x)
		case 2:
			// Y
			y, err := strconv.ParseFloat(part, 32)
			if err != nil {
				return nil, fmt.Errorf("error in y strconv.ParseFloat for %s: %w", part, err)
			}
			workingPart.Y = float32(y)
			// Roll it over
			instructions = append(instructions, workingPart)
			workingPart = SVGDrawInstruction{}
		}
	}
	return instructions, nil
}
