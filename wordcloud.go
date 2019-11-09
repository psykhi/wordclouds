package wordclouds

import (
	"fmt"
	"github.com/fogleman/gg"
	"golang.org/x/image/font"
	"image"
	"image/color"
	"math"
	"math/rand"
	"runtime"
	"sort"
	"sync"
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
	circles         map[float64]*circle
	fonts           map[float64]font.Face
	radii           []float64
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
	BackgroundColor: color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
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

	radius := 1.0
	maxRadius := math.Sqrt(float64(opts.Width*opts.Width + opts.Height*opts.Height))
	circles := make(map[float64]*circle)
	radii := make([]float64, 0)
	for radius < maxRadius {
		circles[radius] = newCircle(float64(opts.Width/2), float64(opts.Height/2), radius, 512)
		radii = append(radii, radius)
		radius = radius + 5.0
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
		circles:         circles,
		fonts:           make(map[float64]font.Face),
		radii:           radii,
	}
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

func (w *Wordcloud) setFont(size float64) {
	_, ok := w.fonts[size]

	if !ok {
		f, err := gg.LoadFontFace(w.opts.FontFile, size)
		if err != nil {
			panic(err)
		}
		w.fonts[size] = f
	}
	w.dc.SetFontFace(w.fonts[size])
}

func (w *Wordcloud) Place(wc WordCount) bool {
	c := w.opts.Colors[rand.Intn(len(w.opts.Colors))]
	w.dc.SetColor(c)

	size := float64(w.opts.FontMaxSize) * (float64(wc.count) / float64(w.sortedWordList[0].count))

	if size < float64(w.opts.FontMinSize) {
		size = float64(w.opts.FontMinSize)
	}
	w.setFont(size)
	width, height := w.dc.MeasureString(wc.word)

	width += 5
	height += 5
	x, y, space := w.nextPos(width, height)
	if !space {
		// fmt.Printf("(%d/%d) Could not place word %s\n", i, len(w.sortedWordList), wc.word)
		return false
	}
	w.dc.DrawStringAnchored(wc.word, x, y, 0.5, 0.5)

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
	return true
}

func (w *Wordcloud) Draw() image.Image {
	consecutiveMisses := 0
	for _, wc := range w.sortedWordList {
		success := w.Place(wc)
		if !success {
			consecutiveMisses++
			if consecutiveMisses > 10 {
				return w.dc.Image()
			}
			continue
		}
		consecutiveMisses = 0
	}
	return w.dc.Image()
}

type workerData struct {
	radius    float64
	positions []point
	width     float64
	height    float64
}

func (w *Wordcloud) nextPos(width float64, height float64) (x float64, y float64, space bool) {
	searching := true
	space = false

	var box Box
	if w.randomPlacement {
		tries := 0
		for searching && tries < 500000 {
			tries++
			x, y = float64(rand.Intn(w.dc.Width())), float64(rand.Intn(w.dc.Height()))
			// Is that position available?
			box.top = y + height/2
			box.left = x - width/2
			box.right = x + width/2
			box.bottom = y - height/2

			w.dc.DrawRectangle(box.x(), box.y(), box.w(), box.h())
			w.dc.Stroke()

			if !box.fits(w.width, w.height) {
				continue
			}
			colliding, _ := w.grid.TestCollision(&box, func(a *Box, b *Box) bool {
				return a.overlaps(b)
			})

			if !colliding {
				space = true
				searching = false
				return
			}
		}
		return
	}

	x, y = w.width, w.height

	resCh := make(chan res, 10000)
	dataCh := make(chan workerData, runtime.NumCPU())
	results := make(map[float64]res)
	done := make(map[float64]bool)
	stopChannels := make([]chan struct{}, 0)
	wg := sync.WaitGroup{}

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		stopCh := make(chan struct{}, 1)
		go func(ch chan struct{}) {
			defer wg.Done()
			for {
				select {
				case d, ok := <-dataCh:
					//fmt.Println("data received")
					if !ok {
						//fmt.Println("bye bye")
						return
					}
					w.testRadius(d.radius, d.positions, d.width, d.height, resCh)
				//fmt.Println("yo")
				case <-ch:
					//fmt.Println("bye")
					return
				}
			}
		}(stopCh)
		stopChannels = append(stopChannels, stopCh)
	}

	go func() {
		for _, r := range w.radii {
			c := w.circles[r]
			//fmt.Println("will start")
			dataCh <- workerData{
				radius:    r,
				positions: c.positions(),
				width:     width,
				height:    height,
			}
		}
		//fmt.Println("closing")
		close(dataCh)
	}()

	for d := range resCh {
		//fmt.Println("received", d)
		results[d.radius] = d
		done[d.radius] = true
		//check if we need to continue
		failed := true
		for _, r := range w.radii {
			if !done[r] {
				//fmt.Println("not done!", r)
				failed = false
				break
			}
			if !results[r].failed {
				//fmt.Println("wooo")
				for _, c := range stopChannels {
					c <- struct{}{}
				}
				//fmt.Println("now waiting")
				wg.Wait()
				return results[r].x, results[r].y, true
			}
		}

		//fmt.Println("out")
		// We tried it all
		if failed {
			for _, c := range stopChannels {
				c <- struct{}{}
			}
			//fmt.Println("Failed to place")
			wg.Wait()
			return
		}

		//fmt.Println(d)
	}
	//fmt.Println("NOOOOO")
	for _, c := range stopChannels {
		c <- struct{}{}
	}
	wg.Wait()
	return
}

type res struct {
	radius float64
	x      float64
	y      float64
	failed bool
}

func (w *Wordcloud) testRadius(radius float64, points []point, width float64, height float64, ch chan res) {
	var box Box
	var x, y float64

	//fmt.Println("starting", radius)
	for _, p := range points {
		y = p.y
		x = p.x

		// Is that position available?

		box.top = y + height/2
		box.left = x - width/2
		box.right = x + width/2
		box.bottom = y - height/2

		//w.dc.DrawRectangle(box.x(), box.y(), box.w(), box.h())
		//w.dc.Stroke()

		if !box.fits(w.width, w.height) {
			continue
		}
		colliding, _ := w.grid.TestCollision(&box, func(a *Box, b *Box) bool {
			return a.overlaps(b)
		})

		if !colliding {
			//space = true
			//fmt.Println("ok", radius)
			ch <- res{
				x:      x,
				y:      y,
				failed: false,
				radius: radius,
			}
			return
		}
	}
	//fmt.Println("nope", radius)
	ch <- res{
		x:      x,
		y:      y,
		failed: true,
		radius: radius,
	}
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
