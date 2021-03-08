package main

import (
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {

	var source string
	var output string
	var shadow int

	app := &cli.App{
		Name:    "Go Image Filter",
		Usage:   "A basic image filtering app",
		Version: "1.0.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "source",
				Aliases:     []string{"src"},
				Usage:       "The source image to transform",
				Required:    true,
				Destination: &source,
			},
			&cli.StringFlag{
				Name:        "output",
				Aliases:     []string{"out"},
				Usage:       "Output path for generated file (png)",
				Required:    true,
				Destination: &output,
			},
		},
		Commands: []*cli.Command{
			{
				Name:      "row-filter",
				Aliases:   []string{"row"},
				Usage:     "Generate an output image where each row is the average, min or max value of all the pixels in source row",
				ArgsUsage: "[mode] (avg|min|max)",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:        "shadow",
						Usage:       "Shadow mask, do not alter pixels with saturation below this threshold (0-255)",
						Value:       0,
						Destination: &shadow,
					},
				},
				Action: func(c *cli.Context) error {
					var mode string
					if c.NArg() > 0 {
						mode = c.Args().Get(0)
					} else {
						mode = "avg"
					}
					rowFilter(source, output, mode, shadow)
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

//
// common
//

type Pixel struct {
	R int
	G int
	B int
	A int
	X int
	Y int
}

func findMinValue(values []int) int {
	min := values[0]
	for _, value := range values {
		if value < min {
			min = value
		}
	}
	return min
}

func findMaxValue(values []int) int {
	max := values[0]
	for _, value := range values {
		if value > max {
			max = value
		}
	}
	return max
}

func rgbaToPixel(r uint32, g uint32, b uint32, a uint32, x int, y int) Pixel {
	return Pixel{
		R: int(r / 257),
		G: int(g / 257),
		B: int(b / 257),
		A: int(a / 257),
		X: x,
		Y: y,
	}
}

func load(srcPath string) ([][]Pixel, int, int) {
	file, _ := os.Open(srcPath)
	defer file.Close()
	img, _, err := image.Decode(file)

	if err != nil {
		log.Fatal(err)
	}

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	var pixels [][]Pixel
	for y := 0; y < height; y++ {
		var row []Pixel
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			px := rgbaToPixel(r, g, b, a, x, y)
			row = append(row, px)
		}
		pixels = append(pixels, row)
	}

	return pixels, height, width
}

//
// filters
//

func rowFilter(srcPath string, outPath string, mode string, shadowMask int) {
	pixels, height, width := load(srcPath)

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	outImg := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	for _, row := range pixels {
		redValues := []int{}
		greenValues := []int{}
		blueValues := []int{}

		avgRed := 0
		avgGreen := 0
		avgBlue := 0

		for _, w := range row {
			redValues = append(redValues, w.R)
			greenValues = append(greenValues, w.G)
			blueValues = append(blueValues, w.B)

			if mode == "avg" {
				avgRed += w.R
				avgGreen += w.G
				avgBlue += w.B
			}

		}

		var rValue int
		var gValue int
		var bValue int

		if mode == "avg" {
			rValue = avgRed / len(row)
			gValue = avgGreen / len(row)
			bValue = avgBlue / len(row)
		} else if mode == "min" {
			rValue = findMinValue(redValues)
			gValue = findMinValue(greenValues)
			bValue = findMinValue(blueValues)
		} else if mode == "max" {
			rValue = findMaxValue(redValues)
			gValue = findMaxValue(greenValues)
			bValue = findMaxValue(blueValues)
		} else {
			log.Fatal("Unknown mode")
		}

		for _, w := range row {
			y, _, _ := color.RGBToYCbCr(uint8(w.R), uint8(w.G), uint8(w.B))
			if y < uint8(shadowMask) {
				outImg.SetRGBA(w.X, w.Y, color.RGBA{uint8(w.R), uint8(w.G), uint8(w.B), 255})
			} else {
				outImg.SetRGBA(w.X, w.Y, color.RGBA{uint8(rValue), uint8(gValue), uint8(bValue), 255})
			}

		}
	}

	f, err := os.Create(outPath)
	if err != nil {
		log.Fatal(err)
	}
	png.Encode(f, outImg)
}
