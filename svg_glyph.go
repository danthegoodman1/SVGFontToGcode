package main

import (
	"errors"
	"fmt"
)

type (
	SVGGlyph struct {
		Char rune
		// How much to advance on the X-axis after drawing the character
		XAdv float64
		D    []SVGDrawInstruction
	}

	SVGDrawInstruction struct {
		Instruction SVGInstruction
		X           float64
		Y           float64
	}

	SVGInstruction rune
)

var (
	// Pen up, move to location
	SVGInstruction_MoveAbsolute SVGInstruction = 'M'
	// Pen down, move to locaiton
	SVGInstruction_LineAbsolute SVGInstruction = 'L'

	ErrUnknownInstruction = errors.New("unknown instruction")
)

// ToGcodeInstructions turns the glyph into a set of Gcode instructions, including the horizontal advance after writing.
// It also returns the new x and y offsets after writing the character.
func (glyph *SVGGlyph) ToGcodeInstructions(initialX, initialY float64) ([]GcodeInstruction, float64, float64, error) {
	var instructions []GcodeInstruction
	for _, svgDrawInstruction := range glyph.D {
		instr := GcodeInstruction{
			Instruction: GcodeLinearMove,
			X:           initialX + svgDrawInstruction.X,
			Y:           initialY + svgDrawInstruction.Y,
		}
		switch svgDrawInstruction.Instruction {
		case SVGInstruction_LineAbsolute:
			instr.Extrude = 1
		case SVGInstruction_MoveAbsolute:
			instr.Extrude = 0
		default:
			return nil, 0, 0, fmt.Errorf("error with instruction %s: %w", svgDrawInstruction.Instruction, ErrUnknownInstruction)
		}
		instructions = append(instructions, instr)
	}

	initialX += glyph.XAdv

	return instructions, initialX, initialY, nil
}
