package main

import (
	"fmt"
	"os"

	"github.com/adeelkhan/code_diff/internal/parser"
)

func main() {
	input := `+func main() {
+	fmt.Println("Hello")
- fmt.Println("World")
}`

	p := parser.New()
	result, err := p.Parse(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Diff Analysis Results")
	fmt.Println("======================")
	fmt.Printf("Additions: %d\n", result.Stats.Additions)
	fmt.Printf("Removals: %d\n", result.Stats.Removals)
	fmt.Printf("Unchanged: %d\n", result.Stats.Unchanged)
	fmt.Println("\nChanges:")
	for _, c := range result.Changes {
		fmt.Printf("  Line %d: %s (%s)\n", c.Position, c.Line, c.Type)
	}
}
