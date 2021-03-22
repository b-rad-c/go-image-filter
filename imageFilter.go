package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/urfave/cli/v2"
)

func main() {

	var source string
	var output string
	var shadow int
	var highlight int
	var checkerBoxSize int

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
				Usage:       "Output path for generated file (png), if not supplied one will be generated based on [source] and filter options",
				Required:    false,
				Destination: &output,
			},
			&cli.IntFlag{
				Name:        "shadow",
				Aliases:     []string{"low"},
				Usage:       "Shadow mask, do not alter pixels with lumanance below this threshold (0-255), default=0 (off)",
				Value:       0,
				Destination: &shadow,
			},
			&cli.IntFlag{
				Name:        "highlight",
				Aliases:     []string{"high"},
				Usage:       "Highlight mask, do not alter pixels with lumanance above this threshold (0-255), default=255 (off)",
				Value:       255,
				Destination: &highlight,
			},
		},
		Commands: []*cli.Command{
			{
				Name:      "row-filter",
				Aliases:   []string{"row"},
				Usage:     "Generate an output image where each row is the average, min or max value of all the pixels in source row",
				ArgsUsage: "[mode] (avg|min|max)",
				Action: func(c *cli.Context) error {
					var mode string
					if c.NArg() > 0 {
						mode = c.Args().Get(0)
					} else {
						mode = "avg"
					}
					rowFilter(source, output, mode, shadow, highlight)
					return nil
				},
			},
			{
				Name:      "checkerbox-filter",
				Aliases:   []string{"checkerbox", "check"},
				Usage:     "Create a checkboard where each box is the average, min or max of the original pixel values in the box",
				ArgsUsage: "[mode] (avg|min|max)",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:        "size",
						Usage:       "Dimension of checkboxes in pixels, default=100",
						Value:       100,
						Destination: &checkerBoxSize,
					},
				},
				Action: func(c *cli.Context) error {
					var mode string
					if c.NArg() > 0 {
						mode = c.Args().Get(0)
					} else {
						mode = "avg"
					}
					checkerboxFilter(source, output, mode, checkerBoxSize, shadow, highlight)
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

func findMinValue(values []uint32) uint32 {
	min := values[0]
	for _, value := range values {
		if value < min {
			min = value
		}
	}
	return min
}

func findMaxValue(values []uint32) uint32 {
	max := values[0]
	for _, value := range values {
		if value > max {
			max = value
		}
	}
	return max
}

func load(srcPath string) (image.Image, int, int) {
	file, _ := os.Open(srcPath)
	defer file.Close()
	img, _, err := image.Decode(file)

	if err != nil {
		log.Fatal(err)
	}

	bounds := img.Bounds()
	height, width := bounds.Max.Y, bounds.Max.X

	return img, height, width
}

func save(outPath string, outImg image.Image) {
	f, err := os.Create(outPath)
	if err != nil {
		log.Fatal(err)
	}
	png.Encode(f, outImg)
}

//
// filters
//

func checkerboxFilter(srcPath string, outPath string, mode string, size int, shadowMask int, highlightMask int) {
	img, height, width := load(srcPath)

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	outImg := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	numRows := height / size
	numColumns := width / size

	for row := 0; row < numRows; row++ {

		for col := 0; col < numColumns; col++ {
			redValues := []uint32{}
			greenValues := []uint32{}
			blueValues := []uint32{}

			avgRed := uint32(0)
			avgGreen := uint32(0)
			avgBlue := uint32(0)

			xOffset := col * size
			xMax := xOffset + size
			yOffset := row * size
			yMax := yOffset + size

			// loop over pixels to get statistis
			for x := xOffset; x < xMax; x++ {
				for y := yOffset; y < yMax; y++ {
					r, g, b, _ := img.At(x, y).RGBA()

					redValues = append(redValues, r)
					greenValues = append(greenValues, g)
					blueValues = append(blueValues, b)

					if mode == "avg" {
						avgRed += r
						avgGreen += g
						avgBlue += b
					}

				}
			}

			// process
			var rValue uint32
			var gValue uint32
			var bValue uint32

			if mode == "avg" {
				rValue = avgRed / uint32(width)
				gValue = avgGreen / uint32(width)
				bValue = avgBlue / uint32(width)
			} else if mode == "min" {
				rValue = findMinValue(redValues)
				gValue = findMinValue(greenValues)
				bValue = findMinValue(blueValues)
			} else if mode == "max" {
				rValue = findMaxValue(redValues)
				gValue = findMaxValue(greenValues)
				bValue = findMaxValue(blueValues)
			} else if mode == "sort" {
				sort.Slice(redValues, func(i, j int) bool { return redValues[i] < redValues[j] })
				sort.Slice(greenValues, func(i, j int) bool { return greenValues[i] < greenValues[j] })
				sort.Slice(blueValues, func(i, j int) bool { return blueValues[i] < blueValues[j] })
			} else {
				log.Fatal("Unknown mode")
			}

			// loop over pixels again to render to values
			sortOffset := 0
			for x := xOffset; x < xMax; x++ {
				for y := yOffset; y < yMax; y++ {
					r, g, b, _ := img.At(x, y).RGBA()
					lum, _, _ := color.RGBToYCbCr(uint8(r/257), uint8(g/257), uint8(b/257))

					if lum < uint8(shadowMask) || lum > uint8(highlightMask) {
						outImg.SetRGBA(x, y, color.RGBA{uint8(r / 257), uint8(g / 257), uint8(b / 257), 255})
					} else if mode == "sort" {
						outImg.SetRGBA(x, y, color.RGBA{uint8(redValues[sortOffset] / 257), uint8(greenValues[sortOffset] / 257), uint8(blueValues[sortOffset] / 257), 255})
						sortOffset++
					} else {
						outImg.SetRGBA(x, y, color.RGBA{uint8(rValue / 257), uint8(gValue / 257), uint8(bValue / 257), 255})
					}
				}
			}

		}

	}

	if outPath == "" {
		var extension = filepath.Ext(srcPath)
		var name = srcPath[0 : len(srcPath)-len(extension)]
		outPath = fmt.Sprint(name, "-checker-", size, "-", mode, "-high-", highlightMask, "-low-", shadowMask, ".png")
	}

	save(outPath, outImg)

}

func rowFilter(srcPath string, outPath string, mode string, shadowMask int, highlightMask int) {
	img, height, width := load(srcPath)

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	outImg := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	for y := 0; y < height; y++ {
		redValues := []uint32{}
		greenValues := []uint32{}
		blueValues := []uint32{}

		avgRed := uint32(0)
		avgGreen := uint32(0)
		avgBlue := uint32(0)

		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()

			redValues = append(redValues, r)
			greenValues = append(greenValues, g)
			blueValues = append(blueValues, b)

			if mode == "avg" {
				avgRed += r
				avgGreen += g
				avgBlue += b
			}

		}

		var rValue uint32
		var gValue uint32
		var bValue uint32

		if mode == "avg" {
			rValue = avgRed / uint32(width)
			gValue = avgGreen / uint32(width)
			bValue = avgBlue / uint32(width)
		} else if mode == "min" {
			rValue = findMinValue(redValues)
			gValue = findMinValue(greenValues)
			bValue = findMinValue(blueValues)
		} else if mode == "max" {
			rValue = findMaxValue(redValues)
			gValue = findMaxValue(greenValues)
			bValue = findMaxValue(blueValues)
		} else if mode == "sort" {
			sort.Slice(redValues, func(i, j int) bool { return redValues[i] < redValues[j] })
			sort.Slice(greenValues, func(i, j int) bool { return greenValues[i] < greenValues[j] })
			sort.Slice(blueValues, func(i, j int) bool { return blueValues[i] < blueValues[j] })
		} else {
			log.Fatal("Unknown mode")
		}

		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			lum, _, _ := color.RGBToYCbCr(uint8(r/257), uint8(g/257), uint8(b/257))

			if lum < uint8(shadowMask) || lum > uint8(highlightMask) {
				outImg.SetRGBA(x, y, color.RGBA{uint8(r / 257), uint8(g / 257), uint8(b / 257), 255})
			} else if mode == "sort" {
				outImg.SetRGBA(x, y, color.RGBA{uint8(redValues[x] / 257), uint8(greenValues[x] / 257), uint8(blueValues[x] / 257), 255})
			} else {
				outImg.SetRGBA(x, y, color.RGBA{uint8(rValue / 257), uint8(gValue / 257), uint8(bValue / 257), 255})
			}

		}
	}

	if outPath == "" {
		var extension = filepath.Ext(srcPath)
		var name = srcPath[0 : len(srcPath)-len(extension)]
		outPath = fmt.Sprint(name, "-row-", mode, "-high-", highlightMask, "-low-", shadowMask, ".png")
	}

	save(outPath, outImg)
}
