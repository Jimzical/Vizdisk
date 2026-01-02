package disktree

// D3Node represents the structure D3.js expects
type D3Node struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Value    int64     `json:"value"` // Size in bytes
	Type     string    `json:"type"`  // "file" or "directory"
	Children []*D3Node `json:"children,omitempty"`
}

func ParseNode(raw any, parentPath string) *D3Node {
	switch v := raw.(type) {
	case []any:
		if len(v) > 0 {
			return parseDirectory(v, parentPath)
		}
	case map[string]any:
		return parseFile(v, parentPath)
	}
	return nil
}

func parseDirectory(list []any, parentPath string) *D3Node {
	meta, ok := list[0].(map[string]any)
	if !ok {
		return nil
	}

	name, _ := meta["name"].(string)
	currentPath := buildPath(parentPath, name)

	node := &D3Node{
		Name: name,
		Path: currentPath,
		Type: "directory",
	}

	for _, childRaw := range list[1:] {
		if childNode := ParseNode(childRaw, currentPath); childNode != nil {
			node.Children = append(node.Children, childNode)
			node.Value += childNode.Value
		}
	}
	return node
}

func parseFile(meta map[string]any, parentPath string) *D3Node {
	name, _ := meta["name"].(string)
	size, _ := meta["asize"].(float64)

	return &D3Node{
		Name:  name,
		Path:  buildPath(parentPath, name),
		Value: int64(size),
		Type:  "file",
	}
}

func buildPath(parent, name string) string {
	if parent == "" {
		return name
	}
	return parent + "/" + name
}
