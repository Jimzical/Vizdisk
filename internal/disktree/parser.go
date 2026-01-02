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
	// Case 1: Directory (Array) -> [ {metadata}, child1, child2... ]
	if list, ok := raw.([]any); ok && len(list) > 0 {
		meta, _ := list[0].(map[string]any)
		name := meta["name"].(string)
		currentPath := parentPath + "/" + name
		if parentPath == "" {
			currentPath = name // Root case
		}

		node := &D3Node{
			Name: name,
			Path: currentPath,
			Type: "directory",
		}

		// Process children (items 1 to end)
		var totalSize int64 = 0
		for _, childRaw := range list[1:] {
			childNode := ParseNode(childRaw, currentPath)
			if childNode != nil {
				node.Children = append(node.Children, childNode)
				totalSize += childNode.Value
			}
		}
		node.Value = totalSize
		return node
	}

	// Case 2: File (Object) -> { "name": "foo", "asize": 123 ... }
	if meta, ok := raw.(map[string]interface{}); ok {
		name := meta["name"].(string)
		size := int64(0)
		if s, ok := meta["asize"].(float64); ok {
			size = int64(s)
		}

		return &D3Node{
			Name:  name,
			Path:  parentPath + "/" + name,
			Value: size,
			Type:  "file",
		}
	}

	return nil
}
