package wordclouds

type naiveGrid struct {
	list []*Box
}

func (g *naiveGrid) Candidates(b *Box) []*Box {
	return g.list
}

func (g *naiveGrid) Add(b *Box) {
	g.list = append(g.list, b)
}

func NewNaiveGrid() Grid {
	return &naiveGrid{make([]*Box, 0)}
}
func (g *naiveGrid) add(b *Box) {
	g.list = append(g.list, b)
}
