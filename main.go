package main

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	log.Println("starting SVG font to Gcode")

	// I know it's deprecated but I am lazy
	//fileNames := make([]string, 0)
	//filepath.Walk("fonts", func(path string, info fs.FileInfo, err error) error {
	//	if !info.IsDir() {
	//		fmt.Println("found file", path)
	//		fileNames = append(fileNames, path)
	//		svgFont, err := parseSVGFont(path)
	//		if err != nil {
	//			return fmt.Errorf("error in parseSVGFont for %s: %w", path, err)
	//		}
	//		fmt.Printf("%+v\n", svgFont)
	//		// FIXME: REMOVE THIS, for development
	//		os.Exit(0)
	//	}
	//	return nil
	//})

	fontName := "HersheySans1.svg"
	textInput := "Hey!"

	svgFont, err := parseSVGFont("fonts/" + fontName)
	if err != nil {
		log.Fatal("error parsing font", err)
	}

	textGlyphs := svgFont.GetCharsForString(textInput)
	fmt.Println("got glyphs", len(textGlyphs))
	for _, glyph := range textGlyphs {
		if glyph != nil {
			fmt.Println(string(glyph.Char))
		} else {
			fmt.Println(" ")
		}
	}

	xInitialOffset := 0.0
	yInitialOffset := 0.0

	xOffset := 0.0 + xInitialOffset
	yOffset := 0.0 + yInitialOffset

	var gCodes []GcodeInstruction
	for _, glyph := range textGlyphs {
		var insts []GcodeInstruction
		var err error
		insts, xOffset, yOffset, err = glyph.ToGcodeInstructions(xOffset, yOffset)
		if err != nil {
			log.Fatal("error parsing gcode instruction for", glyph, err)
		}
		gCodes = append(gCodes, insts...)
	}

	for _, code := range gCodes {
		fmt.Println(code.String())
	}
}

func parseSVGFont(filePath string) (*FontSVG, error) {
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
			x, err := strconv.ParseFloat(xAdv, 64)
			if err != nil {
				// This is lazy but I don't care right now
				panic(fmt.Errorf("error in strconv.ParseInt: %w", err))
			}

			// Parse out the drawing instructions
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
				XAdv: x,
				D:    instructions,
			})
		}
	})

	missingGlyph := doc.Find("missing-glyph")
	if xAdv, exists := missingGlyph.First().Attr("horiz-adv-x"); exists {
		x, err := strconv.ParseFloat(xAdv, 64)
		if err != nil {
			// This is lazy but I don't care right now
			panic(fmt.Errorf("error in strconv.ParseInt: %w", err))
		}
		fontSVG.MissingGlyphXAdv = float64(x)
	}

	return fontSVG, nil
}

func ParseSVGDrawingInstructions(raw string) ([]SVGDrawInstruction, error) {
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, " ")

	// Each part is always format `{instruction} {X} {Y}`, separated by spaces
	var instructions = []SVGDrawInstruction{}
	workingPart := SVGDrawInstruction{}
	for i, part := range parts {
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
			x, err := strconv.ParseFloat(part, 64)
			if err != nil {
				return nil, fmt.Errorf("error in x strconv.ParseFloat for %s: %w", part, err)
			}
			workingPart.X = x
		case 2:
			// Y
			y, err := strconv.ParseFloat(part, 64)
			if err != nil {
				return nil, fmt.Errorf("error in y strconv.ParseFloat for %s: %w", part, err)
			}
			workingPart.Y = y
			// Roll it over
			instructions = append(instructions, workingPart)
			workingPart = SVGDrawInstruction{}
		}
	}
	return instructions, nil
}
