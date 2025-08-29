package hg

type SkatePathGroup struct {
	Paths []*SkatePath
}

// Add adds a SkatePath to the group.
func (g *SkatePathGroup) Add(path *SkatePath) {
	g.Paths = append(g.Paths, path)
}
