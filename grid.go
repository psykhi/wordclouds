package wordclouds

type Grid interface {
	Candidates(b *Box) []*Box
	Add(b *Box)
}
