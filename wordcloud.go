package wordclouds

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"sort"

	"github.com/fogleman/gg"
)

type WordCount struct {
	word  string
	count int
}

type Word2D struct {
	WordCount
	x           float64
	y           float64
	height      float64
	boundingBox *Box
}

type Box struct {
	top    float64
	left   float64
	right  float64
	bottom float64
}

func (b *Box) x() float64 {
	return b.left
}

func (b *Box) y() float64 {
	return b.bottom
}

func (b *Box) w() float64 {
	return b.right - b.left
}

func (b *Box) h() float64 {
	return b.top - b.bottom
}

type Wordcloud struct {
	wordList        map[string]int
	sortedWordList  []WordCount
	grid            *spatialHashMap
	dc              *gg.Context
	overlapCount    int
	words2D         []*Word2D
	availableColors []color.Color
	randomPlacement bool
	width           float64
	height          float64
	opts            Options
}

type Options struct {
	FontMaxSize     int
	FontMinSize     int
	RandomPlacement bool
	FontFile        string
	Colors          []color.Color
	BackgroundColor color.Color
	Width           int
	Height          int
	mask            []*Box
}

var DefaultOptions = Options{
	FontMaxSize:     500,
	FontMinSize:     10,
	RandomPlacement: false,
	FontFile:        "",
	Colors:          []color.Color{color.RGBA{0, 0, 0, 0}},
	BackgroundColor: color.RGBA{0xff, 0xff, 0xff, 0xff},
	Width:           2048,
	Height:          2048,
	mask:            make([]*Box, 0),
}

type Option func(*Options)

func FontFile(path string) Option {
	return func(options *Options) {
		options.FontFile = path
	}
}
func Colors(colors []color.Color) Option {
	return func(options *Options) {
		options.Colors = colors
	}
}

func FontMaxSize(max int) Option {
	return func(options *Options) {
		options.FontMaxSize = max
	}
}
func FontMinSize(min int) Option {
	return func(options *Options) {
		options.FontMinSize = min
	}
}

func MaskBoxes(mask []*Box) Option {
	return func(options *Options) {
		options.mask = mask
	}
}
func Width(w int) Option {
	return func(options *Options) {
		options.Width = w
	}
}

func Height(h int) Option {
	return func(options *Options) {
		options.Height = h
	}
}
func RandomPlacement() Option {
	return func(options *Options) {
		options.RandomPlacement = true
	}
}

func NewWordcloud(wordList map[string]int, options ...Option) *Wordcloud {

	opts := DefaultOptions
	for _, opt := range options {
		opt(&opts)
	}

	sortedWordList := make([]WordCount, 0, len(wordList))
	for word, count := range wordList {
		sortedWordList = append(sortedWordList, WordCount{word: word, count: count})
	}
	sort.Slice(sortedWordList, func(i, j int) bool {
		return sortedWordList[i].count > sortedWordList[j].count
	})
	//sortedWordList = sortedWordList[:5]

	dc := gg.NewContext(opts.Width, opts.Height)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)
	grid := newSpatialHashMap(float64(opts.Width), float64(opts.Height), opts.Height/10)

	for _, b := range opts.mask {
		//dc.DrawRectangle(b.x(), b.y(), b.w(), b.h())
		//dc.Stroke()
		grid.Add(b)
	}
	return &Wordcloud{
		wordList:        wordList,
		sortedWordList:  sortedWordList,
		grid:            grid,
		dc:              dc,
		words2D:         make([]*Word2D, 0),
		randomPlacement: opts.RandomPlacement,
		width:           float64(opts.Width),
		height:          float64(opts.Height),
		opts:            opts,
	}
}
func (w *Wordcloud) BoundingBoxes() []*Box {
	//return w.grid.All()
	return nil
}

func (w *Wordcloud) getPreciseBoundingBoxes(b *Box) []*Box {
	res := make([]*Box, 0)
	step := 5

	defColor := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	for i := int(math.Floor(b.left)); i < int(b.right); i = i + step {
		for j := int(b.bottom); j < int(b.top); j = j + step {
			//fmt.Println(w.dc.Image().At(i, j))
			if w.dc.Image().At(i, j) != defColor {
				res = append(res, &Box{
					float64(j+step) + 5,
					float64(i) - 5,
					float64(i+step) + 5,
					float64(j) - 5,
				})
			}
		}
	}
	return res
}

func (w *Wordcloud) Draw() image.Image {
	consecutiveMisses := 0
	for _, wc := range w.sortedWordList {
		c := w.opts.Colors[rand.Intn(len(w.opts.Colors))]
		w.dc.SetColor(c)

		size := float64(w.opts.FontMaxSize) * (float64(wc.count) / float64(w.sortedWordList[0].count))
		if err := w.dc.LoadFontFace(w.opts.FontFile, size); err != nil {
			panic(err)
		}
		if size < float64(w.opts.FontMinSize) {
			size = float64(w.opts.FontMinSize)
			if err := w.dc.LoadFontFace(w.opts.FontFile, size); err != nil {
				panic(err)
			}
		}
		width, height := w.dc.MeasureString(wc.word)

		width += 5
		height += 5
		x, y, space, overlaps := w.nextPos(width, height)
		if !space {
			// fmt.Printf("(%d/%d) Could not place word %s\n", i, len(w.sortedWordList), wc.word)
			consecutiveMisses++
			if consecutiveMisses > 10 {
				// fmt.Println("No space left. Done.")
				return w.dc.Image()
			}
			continue
		}
		consecutiveMisses = 0
		w.dc.DrawStringAnchored(wc.word, x, y, 0.5, 0.5)
		w.overlapCount += overlaps

		box := &Box{
			y + height/2 + 0.3*height,
			x - width/2,
			x + width/2,
			math.Max(y-height/2, 0),
		}
		if height > 40 {
			preciseBoxes := w.getPreciseBoundingBoxes(box)
			for _, pb := range preciseBoxes {
				w.grid.Add(pb)
				//w.dc.DrawRectangle(pb.x(), pb.y(), pb.w(), pb.h())
				//w.dc.Stroke()
			}
		} else {
			w.grid.Add(box)
		}

		//

		//placed, overlaps := w.AddWord(wc.word, wc.count)
		//if placed {
		// fmt.Printf("(%d/%d) %s: %d occurences, %d collision tests. x %f y %f h %f\n", i, len(w.sortedWordList), wc.word, wc.count, overlaps, x, y, height)
		//fmt.Printf("Grid: %d boxes\n",)
		//} else {
		//	fmt.Printf("Word %s skipped\n", wc.word)
		//}
	}
	//fmt.Printf("%d overlap count\n", w.overlapCount)
	//fmt.Printf("%d overlap count\n", w.overlapCount)
	return w.dc.Image()
}

func (w *Wordcloud) nextPos(width float64, height float64) (x float64, y float64, space bool, overlaps int) {
	searching := true
	space = false

	if w.randomPlacement {
		tries := 0
		for searching && tries < 500000 {
			tries++
			x, y = float64(rand.Intn(w.dc.Width())), float64(rand.Intn(w.dc.Height()))
			// Is that position available?
			box := &Box{
				y + height/2,
				x - width/2,
				x + width/2,
				y - height/2,
			}
			if !box.fits(w.width, w.height) {
				continue
			}
			colliding, overlapTests := w.grid.TestCollision(box, func(a *Box, b *Box) bool {
				return a.overlaps(b)
			})
			overlaps = overlapTests

			if !colliding {
				space = true
				searching = false
				return
			}
		}
		return
	}

	x, y = w.width, w.height
	radius := 1.0
	maxRadius := math.Sqrt(w.width*w.width + w.height*w.height)

	for searching && radius < maxRadius {
		radius = radius + 5
		c := newCircle(w.width/2, w.height/2, radius, 512)
		pts := c.positions()

		for _, p := range pts {
			y = p.y
			x = p.x

			// Is that position available?
			box := &Box{
				y + height/2,
				x - width/2,
				x + width/2,
				y - height/2,
			}
			if !box.fits(w.width, w.height) {
				continue
			}
			colliding, overlapTests := w.grid.TestCollision(box, func(a *Box, b *Box) bool {
				return a.overlaps(b)
			})
			overlaps = overlapTests

			if !colliding {
				space = true
				searching = false
				return
			}
		}
	}
	return
}

func (b *Box) fits(width float64, height float64) bool {
	return b.bottom > 0 && b.top < height && b.left > 0 && b.right < width
}
func (a *Box) overlaps(b *Box) bool {
	return a.left <= b.right && a.right >= b.left && a.top >= b.bottom && a.bottom <= b.top
}
func (a *Box) overlapsRaw(top float64, left float64, right float64, bottom float64) bool {
	return a.left <= right && a.right >= left && a.top >= bottom && a.bottom <= top
}

func (b *Box) String() string {
	return fmt.Sprintf("[x %f y %f w %f h %f]", b.x(), b.y(), b.w(), b.h())
}
