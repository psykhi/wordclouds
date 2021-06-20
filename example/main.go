package main

import (
	"flag"
	"fmt"
	"image/color"
	"image/png"
	"log"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"

	"github.com/psykhi/wordclouds"
	"gopkg.in/yaml.v2"
)

var path = flag.String("input", "input.yaml", "path to flat YAML like {\"word\":42,...}")
var config = flag.String("config", "config.yaml", "path to config file")
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
	FontMaxSize     int          `yaml:"font_max_size"`
	FontMinSize     int          `yaml:"font_min_size"`
	RandomPlacement bool         `yaml:"random_placement"`
	FontFile        string       `yaml:"font_file"`
	Colors          []color.RGBA `yaml:"colors"`
	Width           int          `yaml:"width"`
	Height          int          `yaml:"height"`
	Mask            MaskConf     `yaml:"mask"`
}

type MaskConf struct {
	File  string     `yaml:"file"`
	Color color.RGBA `yaml:"color"`
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
	content, err := os.ReadFile(*path)
	if err != nil {
		panic(err)
	}
	inputWords := make(map[string]int, 0)
	err = yaml.Unmarshal(content, &inputWords)
	if err != nil {
		panic(err)
	}

	// Load config
	conf := DefaultConf
	content, err = os.ReadFile(*config)
	if err == nil {
		err = yaml.Unmarshal(content, &conf)
		if err != nil {
			fmt.Printf("Failed to decode config, using defaults instead: %s\n", err)
		}
	} else {
		fmt.Println("No config file. Using defaults")
	}
	os.Chdir(filepath.Dir(*config))

	confYaml, err := yaml.Marshal(conf)
	fmt.Printf("Configuration: %s\n", confYaml)
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
