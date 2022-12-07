package main

import (
	"strconv"
	"strings"
)

type (
	GcodeInstruction struct {
		Instruction GcodeInstructionPrefix
		X, Y, Z     float64

		// Controls pen (0 or not)
		Extrude float64
	}

	GcodeInstructionPrefix string
)

var (
	GcodeLinearMove GcodeInstructionPrefix = "G1"
)

func (gc *GcodeInstruction) String() string {
	var s []string
	s = append(s, string(gc.Instruction))
	s = append(s, "X"+strconv.FormatFloat(gc.X, 'f', -1, 64))
	s = append(s, "Y"+strconv.FormatFloat(gc.Y, 'f', -1, 64))
	s = append(s, "Z"+strconv.FormatFloat(gc.Z, 'f', -1, 64))
	s = append(s, "E"+strconv.FormatFloat(gc.Extrude, 'f', -1, 64))
	return strings.Join(s, " ")
}
