package parser

import "testing"

func TestSimpleParse(t *testing.T) {
	node, err := Parse("S(Task1, Task2, Task3)")
	if err != nil {
		t.Errorf("expected nil error %s", err)
	}
	if len(node.Children) != 3 {
		t.Error("unexpected children count")
	}
}
