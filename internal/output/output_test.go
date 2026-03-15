package output

import (
	"testing"
)

func TestNewPrinter(t *testing.T) {
	tests := []struct {
		format string
	}{
		{"table"},
		{"json"},
		{"yaml"},
	}
	for _, tt := range tests {
		p := NewPrinter(tt.format)
		if p.Format != tt.format {
			t.Errorf("expected format %s, got %s", tt.format, p.Format)
		}
	}
}

func TestNewTable(t *testing.T) {
	tbl := NewTable("ID", "NAME", "STATUS")
	if len(tbl.Headers) != 3 {
		t.Errorf("expected 3 headers, got %d", len(tbl.Headers))
	}
	if tbl.Headers[0] != "ID" || tbl.Headers[1] != "NAME" || tbl.Headers[2] != "STATUS" {
		t.Errorf("unexpected headers: %v", tbl.Headers)
	}
}

func TestTableData_AddRow(t *testing.T) {
	tbl := NewTable("A", "B")
	tbl.AddRow("1", "2")
	tbl.AddRow("3", "4")
	if len(tbl.Rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(tbl.Rows))
	}
	if tbl.Rows[0][0] != "1" || tbl.Rows[1][1] != "4" {
		t.Errorf("unexpected row data: %v", tbl.Rows)
	}
}

func TestTableData_EmptyTable(t *testing.T) {
	tbl := NewTable("X")
	if len(tbl.Rows) != 0 {
		t.Error("expected 0 rows for new table")
	}
}
