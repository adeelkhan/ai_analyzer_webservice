package parser

import (
	"fmt"
	"strings"

	"github.com/adeelkhan/code_diff/internal/diff"
)

type Parser struct{}

func New() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(input string) (*diff.DiffResult, error) {
	if input == "" {
		return nil, fmt.Errorf("empty input")
	}

	lines := strings.Split(input, "\n")
	changes := make([]diff.Change, 0, len(lines))

	oldLine := 0
	newLine := 0

	for i, line := range lines {
		if line == "" && i == len(lines)-1 {
			continue
		}

		change := diff.Change{
			Position: i + 1,
		}

		switch {
		case strings.HasPrefix(line, "+"):
			change.Type = diff.Add
			change.Line = strings.TrimPrefix(line, "+")
			change.NewLine = newLine + 1
			newLine++
		case strings.HasPrefix(line, "-"):
			change.Type = diff.Remove
			change.Line = strings.TrimPrefix(line, "-")
			change.OldLine = oldLine + 1
			oldLine++
		case strings.HasPrefix(line, " "), line == "":
			change.Type = diff.Unchanged
			change.Line = strings.TrimPrefix(line, " ")
			change.OldLine = oldLine + 1
			change.NewLine = newLine + 1
			oldLine++
			newLine++
		default:
			change.Type = diff.Modify
			change.Line = line
			change.OldLine = oldLine + 1
			change.NewLine = newLine + 1
			oldLine++
			newLine++
		}

		changes = append(changes, change)
	}

	result := &diff.DiffResult{
		Changes: changes,
	}
	result.Stats.Calculate(changes)

	return result, nil
}
