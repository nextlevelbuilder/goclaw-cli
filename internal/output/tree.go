package output

import (
	"fmt"
	"io"
)

// TreeNode represents a node in a tree structure for ASCII rendering.
type TreeNode struct {
	Name     string
	Children []TreeNode
}

// PrintTree renders a tree with ASCII box-drawing prefixes to w.
// Uses ├─, └─, │  for branches matching standard tree(1) style.
func PrintTree(node TreeNode, w io.Writer, prefix string, isLast bool) {
	connector := "├─ "
	childPrefix := prefix + "│  "
	if isLast {
		connector = "└─ "
		childPrefix = prefix + "   "
	}
	fmt.Fprintf(w, "%s%s%s\n", prefix, connector, node.Name)
	for i, child := range node.Children {
		PrintTree(child, w, childPrefix, i == len(node.Children)-1)
	}
}

// PrintTreeRoot renders a root node (no prefix connector) then its children.
func PrintTreeRoot(node TreeNode, w io.Writer) {
	fmt.Fprintf(w, "%s\n", node.Name)
	for i, child := range node.Children {
		PrintTree(child, w, "", i == len(node.Children)-1)
	}
}

// VaultEntriesToTreeNode converts a flat list of vault tree entries (as []map[string]any)
// into a nested TreeNode for ASCII rendering. Each entry has "name" and "type" keys.
// This is a shallow list (server returns one depth level) — nesting by "/" in name.
func VaultEntriesToTreeNode(root string, entries []map[string]any) TreeNode {
	rootNode := TreeNode{Name: root}
	for _, e := range entries {
		name := ""
		if v, ok := e["name"].(string); ok {
			name = v
		}
		if name == "" {
			if v, ok := e["path"].(string); ok {
				name = v
			}
		}
		if name == "" {
			continue
		}
		rootNode.Children = append(rootNode.Children, TreeNode{Name: name})
	}
	return rootNode
}
