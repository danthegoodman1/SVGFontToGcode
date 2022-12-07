package main

type (
	FontSVG struct {
		Glyphs           []SVGGlyph
		MissingGlyphXAdv float64
	}
)

// GetChar returns nil if not found
func (font *FontSVG) GetChar(char rune) *SVGGlyph {
	for _, glyph := range font.Glyphs {
		if char == glyph.Char {
			return &glyph
		}
	}
	return nil
}

// GetCharsForString returns the array of glyphs as they appear in the string, including duplicates
func (font *FontSVG) GetCharsForString(s string) []*SVGGlyph {
	var glyphs []*SVGGlyph
	runes := []rune(s)
	for _, r := range runes {
		found := false
		for _, glyph := range font.Glyphs {
			if r == glyph.Char {
				glyphs = append(glyphs, &glyph)
				found = true
				break
			}
		}
		if !found {
			glyphs = append(glyphs, nil)
		}
	}
	return glyphs
}
