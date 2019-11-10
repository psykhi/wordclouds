package wordclouds

import (
	"bufio"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"image/color"
	"image/png"
	"os"
	"testing"
	"time"
)

func TestWordcloud_Draw(t *testing.T) {
	colorsRGBA := []color.RGBA{
		{0x1b, 0x1b, 0x1b, 0xff},
		{0x48, 0x48, 0x4B, 0xff},
		{0x59, 0x3a, 0xee, 0xff},
		{0x65, 0xCD, 0xFA, 0xff},
		{0x70, 0xD6, 0xBF, 0xff},
	}
	colors := make([]color.Color, 0)
	for _, c := range colorsRGBA {
		colors = append(colors, c)
	}
	// Load words
	f, err := os.Open("testdata/input.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	dec := json.NewDecoder(reader)
	inputWords := make(map[string]int, 0)
	err = dec.Decode(&inputWords)
	assert.NoError(t, err)

	t0 := time.Now()

	boxes := Mask(
		"testdata/mask.png",
		2048,
		2048,
		color.RGBA{
			R: 0,
			G: 0,
			B: 0,
			A: 0,
		})

	t.Logf("Mask loading took %v", time.Since(t0))
	t0 = time.Now()
	w := NewWordcloud(inputWords,
		FontFile("testdata/Roboto-Regular.ttf"),
		FontMaxSize(300),
		FontMinSize(30),
		Colors(colors),
		MaskBoxes(boxes),
		Height(2048),
		Width(2048))

	t.Logf("Wordcloud init took %v", time.Since(t0))
	t0 = time.Now()

	img := w.Draw()

	t.Logf("Drawing took %v", time.Since(t0))
	t0 = time.Now()

	outputFile, err := os.Create("res.png")
	assert.NoError(t, err)

	// Encode takes a writer interface and an image interface
	// We pass it the File and the RGBA
	png.Encode(outputFile, img)

	// Don't forget to close files
	outputFile.Close()
}
