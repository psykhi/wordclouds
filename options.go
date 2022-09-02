package wordclouds

import (
	"image/color"
	"io"
	"os"
	"path/filepath"
)

type Options struct {
	FontMaxSize     int
	FontMinSize     int
	RandomPlacement bool
	FontFile        []byte
	Colors          []color.Color
	BackgroundColor color.Color
	Width           int
	Height          int
	Mask            []*Box
	SizeFunction    sizeFunction
	Debug           bool
}

var defaultOptions = Options{
	FontMaxSize:     500,
	FontMinSize:     10,
	RandomPlacement: false,
	FontFile:        nil,
	Colors:          []color.Color{color.RGBA{}},
	BackgroundColor: color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
	Width:           2048,
	Height:          2048,
	Mask:            make([]*Box, 0),
	SizeFunction:    sizeLinear,
	Debug:           false,
}

type Option func(*Options)

// Path to font file
func FontFile[T ~string | ~[]byte](font T) Option {
	var b []byte
	// any(font) is work around to aboid compile error
	// with the current generics specification.
	// this might be cleared in the future.
	switch v := any(font).(type) {
	case string:
		f, err := os.Open(filepath.Clean(v))
		if err != nil {
			panic(err)
		}
		defer f.Close()
		b, err = io.ReadAll(f)
		if err != nil {
			panic(err)
		}
	case []byte:
		b = v
	}
	return func(options *Options) {
		options.FontFile = b
	}
}

// Output file background color
func BackgroundColor(color color.Color) Option {
	return func(options *Options) {
		options.BackgroundColor = color
	}
}

// Colors to use for the words
func Colors(colors []color.Color) Option {
	return func(options *Options) {
		options.Colors = colors
	}
}

// Max font size
func FontMaxSize(max int) Option {
	return func(options *Options) {
		options.FontMaxSize = max
	}
}

// Min font size
func FontMinSize(min int) Option {
	return func(options *Options) {
		options.FontMinSize = min
	}
}

// A list of bounding boxes where words can not be placed.
// See Mask
func MaskBoxes(mask []*Box) Option {
	return func(options *Options) {
		options.Mask = mask
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

// Place words randomly
func RandomPlacement(do bool) Option {
	return func(options *Options) {
		options.RandomPlacement = do
	}
}

// Set word font sizing function
func WordSizeFunction(f string) Option {
	return func(options *Options) {
		switch f {
		case SizeFunctionLinear:
			options.SizeFunction = sizeLinear
		case SizeFunctionSqrt:
			options.SizeFunction = sizeSqrt
		case SizeFunctionSqrtInverse:
			options.SizeFunction = sizeSqrtInverse
		default:
			panic("No such size function " + f)
		}
	}
}

// Draw bounding boxes around words
func Debug() Option {
	return func(options *Options) {
		options.Debug = true
	}
}
