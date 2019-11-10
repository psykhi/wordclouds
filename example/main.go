package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/psykhi/wordclouds"
	"image/color"
	"image/png"
	"log"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"
)

var path = flag.String("input", "input.json", "path to flat JSON like {\"word\":42,...}")
var config = flag.String("config", "config.json", "path to config file")
var output = flag.String("output", "output.png", "path to output image")
var cpuprofile = flag.String("cpuprofile", "profile", "write cpu profile to file")

var DefaultColors = []color.RGBA{
	{0x1b, 0x1b, 0x1b, 0xff},
	{0x48, 0x48, 0x4B, 0xff},
	{0x59, 0x3a, 0xee, 0xff},
	{0x65, 0xCD, 0xFA, 0xff},
	{0x70, 0xD6, 0xBF, 0xff},
}

type Conf struct {
	FontMaxSize     int          `json:"font_max_size"`
	FontMinSize     int          `json:"font_min_size"`
	RandomPlacement bool         `json:"random_placement"`
	FontFile        string       `json:"font_file"`
	Colors          []color.RGBA `json:"colors"`
	Width           int          `json:"width"`
	Height          int          `json:"height"`
	Mask            MaskConf     `json:"mask"`
}

type MaskConf struct {
	File  string     `json:"file"`
	Color color.RGBA `json:"color"`
}

var DefaultConf = Conf{
	FontMaxSize:     700,
	FontMinSize:     10,
	RandomPlacement: false,
	FontFile:        "./fonts/roboto/Roboto-Regular.ttf",
	Colors:          DefaultColors,
	Width:           4096,
	Height:          4096,
	Mask: MaskConf{"", color.RGBA{
		R: 0,
		G: 0,
		B: 0,
		A: 0,
	}},
}

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// Load words
	f, err := os.Open(*path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	dec := json.NewDecoder(reader)
	inputWords := make(map[string]int, 0)
	err = dec.Decode(&inputWords)
	if err != nil {
		panic(err)
	}

	// Load config
	conf := DefaultConf
	f, err = os.Open(*config)
	if err == nil {
		defer f.Close()
		reader = bufio.NewReader(f)
		dec = json.NewDecoder(reader)
		err = dec.Decode(&conf)
		if err != nil {
			fmt.Printf("Failed to decode config, using defaults instead: %s\n", err)
		}
	} else {
		fmt.Println("No config file. Using defaults")
	}
	os.Chdir(filepath.Dir(*config))

	confJson, _ := json.Marshal(conf)
	fmt.Printf("Configuration: %s\n", confJson)
	err = json.Unmarshal(confJson, &conf)
	if err != nil {
		fmt.Println(err)
	}

	var boxes []*wordclouds.Box
	if conf.Mask.File != "" {
		boxes = wordclouds.Mask(
			conf.Mask.File,
			conf.Width,
			conf.Height,
			conf.Mask.Color)
	}

	colors := make([]color.Color, 0)
	for _, c := range conf.Colors {
		colors = append(colors, c)
	}

	start := time.Now()
	w := wordclouds.NewWordcloud(inputWords,
		wordclouds.FontFile(conf.FontFile),
		wordclouds.FontMaxSize(conf.FontMaxSize),
		wordclouds.FontMinSize(conf.FontMinSize),
		wordclouds.Colors(colors),
		wordclouds.MaskBoxes(boxes),
		wordclouds.Height(conf.Height),
		wordclouds.Width(conf.Width),
		wordclouds.RandomPlacement(conf.RandomPlacement))

	img := w.Draw()
	outputFile, err := os.Create(*output)
	if err != nil {
		// Handle error
	}

	// Encode takes a writer interface and an image interface
	// We pass it the File and the RGBA
	png.Encode(outputFile, img)

	// Don't forget to close files
	outputFile.Close()
	fmt.Printf("Done in %v\n", time.Since(start))
}
