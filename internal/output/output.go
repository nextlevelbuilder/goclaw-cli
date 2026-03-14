package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Printer handles output formatting based on selected format.
type Printer struct {
	Format string
}

// NewPrinter creates a printer with the given format (table, json, yaml).
func NewPrinter(format string) *Printer {
	return &Printer{Format: format}
}

// Print outputs data in the configured format.
// For table format, pass headers []string and rows [][]string.
// For json/yaml, pass any serializable value.
func (p *Printer) Print(data any) {
	switch p.Format {
	case "json":
		p.printJSON(data)
	case "yaml":
		p.printYAML(data)
	default:
		// Table format requires TableData
		if td, ok := data.(*TableData); ok {
			p.printTable(td)
		} else {
			p.printJSON(data) // fallback for non-table data
		}
	}
}

// TableData holds tabular output data.
type TableData struct {
	Headers []string
	Rows    [][]string
}

// NewTable creates a TableData with the given headers.
func NewTable(headers ...string) *TableData {
	return &TableData{Headers: headers}
}

// AddRow adds a row to the table.
func (td *TableData) AddRow(values ...string) {
	td.Rows = append(td.Rows, values)
}

func (p *Printer) printJSON(data any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(data)
}

func (p *Printer) printYAML(data any) {
	enc := yaml.NewEncoder(os.Stdout)
	enc.SetIndent(2)
	_ = enc.Encode(data)
}

func (p *Printer) printTable(td *TableData) {
	if len(td.Rows) == 0 {
		fmt.Println("No results found.")
		return
	}

	// Calculate column widths
	widths := make([]int, len(td.Headers))
	for i, h := range td.Headers {
		widths[i] = len(h)
	}
	for _, row := range td.Rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	printRow(td.Headers, widths)
	// Print separator
	sep := make([]string, len(widths))
	for i, w := range widths {
		sep[i] = strings.Repeat("-", w)
	}
	fmt.Println(strings.Join(sep, "  "))

	// Print rows
	for _, row := range td.Rows {
		printRow(row, widths)
	}
}

func printRow(cells []string, widths []int) {
	parts := make([]string, len(cells))
	for i, cell := range cells {
		w := 0
		if i < len(widths) {
			w = widths[i]
		}
		parts[i] = fmt.Sprintf("%-*s", w, cell)
	}
	fmt.Println(strings.Join(parts, "  "))
}

// Error prints an error message. In JSON mode, outputs structured error.
func (p *Printer) Error(err error) {
	if p.Format == "json" {
		p.printJSON(map[string]any{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}
	fmt.Fprintf(os.Stderr, "Error: %s\n", err)
}

// Success prints a success message.
func (p *Printer) Success(msg string) {
	if p.Format == "json" {
		p.printJSON(map[string]any{
			"ok":      true,
			"message": msg,
		})
		return
	}
	fmt.Println(msg)
}
