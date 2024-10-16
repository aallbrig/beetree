package tree

type Node struct {
	Parent   *Node
	Children []*Node
	Behavior *Behavior
}

func (n *Node) AddChild(child *Node) {
	child.Parent = n
	n.Children = append(n.Children, child)
}

func NewNode(b *Behavior) *Node {
	n := &Node{
		Behavior: b,
	}
	/*
		switch behavior := b.(type) {
		case Sequence:

		}
	*/
	return n
}
