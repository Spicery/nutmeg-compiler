package common

type Path struct {
	SiblingPosition int // Position among siblings
	Parent          *Node
	Others          *Path
}

func (p *Path) Node() *Node {
	return p.Parent.Children[p.SiblingPosition]
}
