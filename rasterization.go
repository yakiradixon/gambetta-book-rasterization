package main

import (
	"fmt"
	"os"
	"strings"
)

type point struct {
	x int
	y int
}

type color struct {
	r int
	g int
	b int
}

func split(s string) (string, string) {
	i := strings.LastIndex(s[:70], " ")
	return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i:])
}

type canvas struct {
	width  int
	height int
	pixels [][]color
}

func (c canvas) init(w, h int) canvas {
	pixels := make([][]color, h+1)
	for x, _ := range pixels {
		pixels[x] = make([]color, w)
		for y, _ := range pixels[x] {
			pixels[x][y] = color{255, 255, 255}
		}
	}
	return canvas{w, h, pixels}
}

func toPPM(c canvas) string {
	return ppmHeader(c) + ppmPixelData(c) + ppmFooter()
}

func ppmHeader(c canvas) string {
	PPMFlavor := "P3"
	MaxColorValue := 255
	return fmt.Sprintf("%s\n%d %d\n%d\n", PPMFlavor, c.width, c.height, MaxColorValue)
}

func ppmPixelData(c canvas) string {
	var sb strings.Builder
	var psb *strings.Builder = &sb
	var rowdata string
	var row *string = &rowdata
	var data string

	var MaxCharacters = 70

	for _, pixel := range c.pixels {
		for _, color := range pixel {
			writePixelDataFor(psb, color.r, row)
			writePixelDataFor(psb, color.g, row)
			writePixelDataFor(psb, color.b, row)
		}
		l := len(rowdata)
		for l > 0 {
			if l > MaxCharacters {
				d, s := split(rowdata)
				data += d + "\n"
				rowdata = s
				l = len(rowdata)
			} else {
				data += strings.TrimSpace(rowdata) + "\n"
				rowdata = ""
				l = len(rowdata)
			}
		}
	}
	return data
}

func ppmFooter() string {
	return "\n"
}

func writePixelDataFor(psb *strings.Builder, c int, r *string) {
	(*psb).WriteString(fmt.Sprintf("%d ", c))
	*r += (*psb).String()
	(*psb).Reset()
}

func (c color) clamp() color {
	MinColorValue := 0
	MaxColorValue := 255
	return color{clamp(c.r, MinColorValue, MaxColorValue), clamp(c.g, MinColorValue, MaxColorValue), clamp(c.b, MinColorValue, MaxColorValue)}
}

func clamp(x, min, max int) int {
	if x < min {
		x = min
	} else if x > max {
		x = max
	}
	return x
}

func (c canvas) putPixel(x int, y int, color color) {
	sx := (c.width / 2) + x
	sy := (c.height / 2) - y - 1
	c.pixels[sy][sx] = color
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func render(canvases []canvas) {
	for i, c := range canvases {
		ppm := toPPM(c)
		filename := fmt.Sprintf("%s_%d_latest.ppm", "rasterization", i+1)
		f, err := os.Create(filename)
		check(err)
		defer f.Close()
		_, err1 := f.WriteString(ppm)
		check(err1)
		f.Sync()
	}
}

func (c canvas) drawLine(p0 point, p1 point) {
	a := (p1.y - p0.y) / (p1.x - p0.x)
	b := p0.y - a*p0.x
	for x := p0.x; x <= p1.x; x++ {
		y := a*x + b
		c.putPixel(x, y, color{255, 0, 0})
	}
}

func main() {
	c := canvas{}
	c = c.init(200, 200)
	c.drawLine(point{1, 1}, point{50, 50})
	render([]canvas{c})
}
