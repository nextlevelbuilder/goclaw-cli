package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintTreeRoot_NoChildren(t *testing.T) {
	node := TreeNode{Name: "root"}
	var buf bytes.Buffer
	PrintTreeRoot(node, &buf)
	got := buf.String()
	if !strings.Contains(got, "root") {
		t.Errorf("expected root in output, got %q", got)
	}
}

func TestPrintTreeRoot_WithChildren(t *testing.T) {
	node := TreeNode{
		Name: "/",
		Children: []TreeNode{
			{Name: "docs"},
			{Name: "notes"},
		},
	}
	var buf bytes.Buffer
	PrintTreeRoot(node, &buf)
	out := buf.String()

	if !strings.Contains(out, "/") {
		t.Error("expected root '/' in output")
	}
	if !strings.Contains(out, "docs") {
		t.Error("expected 'docs' in output")
	}
	if !strings.Contains(out, "notes") {
		t.Error("expected 'notes' in output")
	}
	// Last child uses └─, others use ├─
	if !strings.Contains(out, "├─") {
		t.Error("expected ├─ connector for non-last child")
	}
	if !strings.Contains(out, "└─") {
		t.Error("expected └─ connector for last child")
	}
}

func TestPrintTree_NestedChildren(t *testing.T) {
	node := TreeNode{
		Name: "parent",
		Children: []TreeNode{
			{
				Name: "child",
				Children: []TreeNode{
					{Name: "grandchild"},
				},
			},
		},
	}
	var buf bytes.Buffer
	PrintTree(node, &buf, "", true)
	out := buf.String()

	if !strings.Contains(out, "parent") {
		t.Error("expected 'parent' in output")
	}
	if !strings.Contains(out, "child") {
		t.Error("expected 'child' in output")
	}
	if !strings.Contains(out, "grandchild") {
		t.Error("expected 'grandchild' in output")
	}
}

func TestVaultEntriesToTreeNode_Empty(t *testing.T) {
	node := VaultEntriesToTreeNode("root", nil)
	if node.Name != "root" {
		t.Errorf("expected name=root, got %q", node.Name)
	}
	if len(node.Children) != 0 {
		t.Errorf("expected no children, got %d", len(node.Children))
	}
}

func TestVaultEntriesToTreeNode_WithEntries(t *testing.T) {
	entries := []map[string]any{
		{"name": "agents/", "type": "folder"},
		{"name": "notes/readme.md", "type": "file"},
		{"path": "fallback.md"}, // no "name" key — falls back to "path"
	}
	node := VaultEntriesToTreeNode("/", entries)
	if len(node.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(node.Children))
	}
	if node.Children[0].Name != "agents/" {
		t.Errorf("expected 'agents/', got %q", node.Children[0].Name)
	}
	if node.Children[2].Name != "fallback.md" {
		t.Errorf("expected 'fallback.md' via path fallback, got %q", node.Children[2].Name)
	}
}

func TestVaultEntriesToTreeNode_SkipsBlankName(t *testing.T) {
	entries := []map[string]any{
		{"name": ""},   // blank name, no path — should be skipped
		{"name": "ok"}, // valid
	}
	node := VaultEntriesToTreeNode("root", entries)
	if len(node.Children) != 1 {
		t.Errorf("expected 1 child (blank skipped), got %d", len(node.Children))
	}
}

func TestPrintTree_IsLastConnectors(t *testing.T) {
	// When isLast=true, use └─ ; when false, use ├─
	node := TreeNode{Name: "node"}
	var buf1, buf2 bytes.Buffer
	PrintTree(node, &buf1, "", true)
	PrintTree(node, &buf2, "", false)

	if !strings.Contains(buf1.String(), "└─") {
		t.Error("isLast=true should use └─")
	}
	if !strings.Contains(buf2.String(), "├─") {
		t.Error("isLast=false should use ├─")
	}
}
