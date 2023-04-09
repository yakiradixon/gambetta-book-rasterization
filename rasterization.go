package main

import (
	"fmt"
	"math"
	"os"
	"strings"
)

type point struct {
	x float64
	y float64
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

func abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}
func swap(p0 *point, p1 *point) {
	p0.x, p1.x = p1.x, p0.x
	p0.y, p1.y = p1.y, p0.y
}

func interpolate(i0, d0, i1, d1 float64) []float64 {
	values := []float64{}
	if i0 == i1 {
		return []float64{d0}
	}
	a := (d1 - d0) / (i1 - i0)
	d := d0
	for i := i0; i <= i1; i++ {
		values = append(values, d)
		d = d + a
	}
	return values
}

func (c canvas) drawLine(p0 point, p1 point, color color) {
	dx := p1.x - p0.x
	dy := p1.y - p0.y

	// If there are more different values for x than y, the line is more horizontal than vertical
	if math.Abs(dx) > math.Abs(dy) {
		if p0.x > p1.x {
			swap(&p0, &p1)
		}
		ys := interpolate(p0.x, p0.y, p1.x, p1.y)
		for x := p0.x; x <= p1.x; x++ {
			c.putPixel(int(x), int(ys[int(x-p0.x)]), color)
		}

	} else {
		if p0.y > p1.y {
			swap(&p0, &p1)
		}
		xs := interpolate(p0.y, p0.x, p1.y, p1.x)
		for y := p0.y; y <= p1.y; y++ {
			c.putPixel(int(xs[int(y-p0.y)]), int(y), color)
		}
	}
}

func (c canvas) drawWireFrameTriangle(p0, p1, p2 point, color color) {
	c.drawLine(p0, p1, color)
	c.drawLine(p1, p2, color)
	c.drawLine(p2, p0, color)
}

func main() {
	c := canvas{}
	c = c.init(600, 600)
	// c.drawLine(point{-200, -100}, point{240, 120}, color{255, 0, 0})
	// c.drawLine(point{-50, -200}, point{60, 240}, color{0, 255, 0})
	c.drawWireFrameTriangle(point{-200, -250}, point{200, 50}, point{20, 250}, color{255, 0, 0})
	render([]canvas{c})
}
