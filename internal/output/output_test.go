package output

import (
	"errors"
	"os"
	"strings"
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

// captureStdoutPrinter redirects os.Stdout, runs fn, restores and returns output.
func captureStdoutPrinter(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	r.Close()
	return string(buf[:n])
}

func TestPrinter_JSON(t *testing.T) {
	p := NewPrinter("json")
	out := captureStdoutPrinter(t, func() {
		p.Print(map[string]string{"key": "value"})
	})
	if !strings.Contains(out, `"key"`) {
		t.Errorf("JSON output missing key: %q", out)
	}
}

func TestPrinter_YAML(t *testing.T) {
	p := NewPrinter("yaml")
	out := captureStdoutPrinter(t, func() {
		p.Print(map[string]string{"name": "test"})
	})
	if !strings.Contains(out, "name: test") {
		t.Errorf("YAML output missing field: %q", out)
	}
}

func TestPrinter_Table_WithRows(t *testing.T) {
	p := NewPrinter("table")
	tbl := NewTable("ID", "NAME")
	tbl.AddRow("1", "alpha")
	tbl.AddRow("2", "beta")
	out := captureStdoutPrinter(t, func() {
		p.Print(tbl)
	})
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") {
		t.Errorf("table output missing rows: %q", out)
	}
}

func TestPrinter_Table_EmptyRows(t *testing.T) {
	p := NewPrinter("table")
	tbl := NewTable("ID", "NAME")
	out := captureStdoutPrinter(t, func() {
		p.Print(tbl)
	})
	if !strings.Contains(out, "No results") {
		t.Errorf("expected 'No results' message, got: %q", out)
	}
}

func TestPrinter_Table_NonTableData_FallsBackToJSON(t *testing.T) {
	p := NewPrinter("table")
	out := captureStdoutPrinter(t, func() {
		p.Print(map[string]string{"foo": "bar"})
	})
	if !strings.Contains(out, "foo") {
		t.Errorf("expected JSON fallback output, got: %q", out)
	}
}

func TestPrinter_Error_JSON(t *testing.T) {
	p := NewPrinter("json")
	out := captureStdoutPrinter(t, func() {
		p.Error(errors.New("something failed"))
	})
	if !strings.Contains(out, "something failed") {
		t.Errorf("error JSON missing message: %q", out)
	}
}

func TestPrinter_Success_JSON(t *testing.T) {
	p := NewPrinter("json")
	out := captureStdoutPrinter(t, func() {
		p.Success("created successfully")
	})
	if !strings.Contains(out, "created successfully") {
		t.Errorf("success JSON missing message: %q", out)
	}
}

func TestPrinter_Success_Table(t *testing.T) {
	p := NewPrinter("table")
	out := captureStdoutPrinter(t, func() {
		p.Success("done")
	})
	if !strings.Contains(out, "done") {
		t.Errorf("success table missing message: %q", out)
	}
}
