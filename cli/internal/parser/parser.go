package parser

import (
	"fmt"
	"strings"
)

type Node struct {
	Type     string
	Children []*Node
}

func Parse(input string) (*Node, error) {
	input = strings.TrimSpace(input)
	if len(input) == 0 {
		return nil, fmt.Errorf("empty input")
	}

	if !strings.Contains(input, "(") {
		return &Node{Type: input}, nil
	}

	openParen := strings.Index(input, "(")
	closeParen := strings.LastIndex(input, ")")

	if openParen == -1 || closeParen == -1 || closeParen < openParen {
		return nil, fmt.Errorf("invalid input: mismatched parentheses")
	}

	nodeType := strings.TrimSpace(input[:openParen])
	childrenStr := input[openParen+1 : closeParen]

	node := &Node{Type: nodeType}

	for _, childStr := range splitChildren(childrenStr) {
		child, err := Parse(childStr)
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, child)
	}

	return node, nil
}

func splitChildren(input string) []string {
	var result []string
	var current string
	parenthesesCount := 0

	for _, char := range input {
		switch char {
		case '(':
			parenthesesCount++
		case ')':
			parenthesesCount--
		case ',':
			if parenthesesCount == 0 {
				result = append(result, strings.TrimSpace(current))
				current = ""
				continue
			}
		}
		current += string(char)
	}

	if current != "" {
		result = append(result, strings.TrimSpace(current))
	}

	return result
}
