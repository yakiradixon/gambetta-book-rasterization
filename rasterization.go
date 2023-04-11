package main

import (
	"fmt"
	"math"
	"os"
	"strings"

	"golang.org/x/exp/slices"
)

type point struct {
	x float64
	y float64
	h float64
}

type color struct {
	r float64
	g float64
	b float64
}

type vertex struct {
	x float64
	y float64
	z float64
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
			writePixelDataFor(psb, int(color.r), row)
			writePixelDataFor(psb, int(color.g), row)
			writePixelDataFor(psb, int(color.b), row)
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
	MinColorValue := 0.0
	MaxColorValue := 255.0
	return color{clamp(c.r, MinColorValue, MaxColorValue), clamp(c.g, MinColorValue, MaxColorValue), clamp(c.b, MinColorValue, MaxColorValue)}
}

func clamp(x, min, max float64) float64 {
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

func (c canvas) drawFilledTriangle(p0, p1, p2 point, color color) {
	if p1.y < p0.y {
		swap(&p1, &p0)
	}
	if p2.y < p0.y {
		swap(&p2, &p0)
	}
	if p2.y < p2.y {
		swap(&p2, &p1)
	}
	x01 := interpolate(p0.y, p0.x, p1.y, p1.x)
	x12 := interpolate(p1.y, p1.x, p2.y, p2.x)
	x02 := interpolate(p0.y, p0.x, p2.y, p2.x)
	x01 = x01[:len(x01)-1]
	x012 := append(x01, x12...)

	m := len(x012) / 2
	xLeft := []float64{}
	xRight := []float64{}
	if x02[m] < x012[m] {
		xLeft = slices.Clone(x02)
		xRight = slices.Clone(x012)
	} else {
		xLeft = slices.Clone(x012)
		xRight = slices.Clone(x02)
	}

	for y := p0.y; y <= p2.y; y++ {
		for x := xLeft[int(y-p0.y)]; x <= xRight[int(y-p0.y)]; x++ {
			c.putPixel(int(x), int(y), color)
		}
	}
}

func (c canvas) drawShadedTriangle(p0, p1, p2 point, color color) {
	if p1.y < p0.y {
		swap(&p1, &p0)
	}
	if p2.y < p0.y {
		swap(&p2, &p0)
	}
	if p2.y < p2.y {
		swap(&p2, &p1)
	}
	x01 := interpolate(p0.y, p0.x, p1.y, p1.x)
	h01 := interpolate(p0.y, p0.h, p1.y, p1.h)
	x12 := interpolate(p1.y, p1.x, p2.y, p2.x)
	h12 := interpolate(p1.y, p1.h, p2.y, p2.h)
	x02 := interpolate(p0.y, p0.x, p2.y, p2.x)
	h02 := interpolate(p0.y, p0.h, p2.y, p2.h)

	x01 = x01[:len(x01)-1]
	x012 := append(x01, x12...)

	h01 = h01[:len(h01)-1]
	h012 := append(h01, h12...)

	m := len(x012) / 2
	xLeft := []float64{}
	xRight := []float64{}
	hLeft := []float64{}
	hRight := []float64{}
	if x02[m] < x012[m] {
		xLeft = slices.Clone(x02)
		xRight = slices.Clone(x012)

		hLeft = slices.Clone(h02)
		hRight = slices.Clone(h012)
	} else {
		xLeft = slices.Clone(x012)
		xRight = slices.Clone(x02)

		hLeft = slices.Clone(h012)
		hRight = slices.Clone(h02)

	}

	for y := p0.y; y <= p2.y; y++ {
		h := interpolate(xLeft[int(y-p0.y)], hLeft[int(y-p0.y)], xRight[int(y-p0.y)], hRight[int(y-p0.y)])
		for x := xLeft[int(y-p0.y)]; x <= xRight[int(y-p0.y)]; x++ {
			c.putPixel(int(x), int(y), multiplyColor(h[int(x-xLeft[int(y-p0.y)])], color))
		}
	}
}

func multiplyColor(v float64, c color) color {
	return color{v * c.r, v * c.g, v * c.b}
}

func (c canvas) toViewport(p point) point {
	viewportSize := 1.0
	return point{x: p.x * float64(c.width) / viewportSize,
		y: p.y * float64(c.height) / viewportSize,
		h: 1.0}
}

func (c canvas) projectVertex(v vertex) point {
	projectionPlaneZ := 1.0
	return c.toViewport(point{x: v.x * projectionPlaneZ / v.z, y: v.y * projectionPlaneZ / v.z, h: 1.0})
}

func main() {
	c := canvas{}
	c = c.init(600, 600)
	// c.drawLine(point{-200, -100}, point{240, 120}, color{255, 0, 0})
	// c.drawLine(point{-50, -200}, point{60, 240}, color{0, 255, 0})
	//c.drawWireFrameTriangle(point{-200, -250}, point{200, 50}, point{20, 250}, color{0, 0, 0})
	// c.drawFilledTriangle(point{-200, -250}, point{200, 50}, point{20, 250}, color{0, 255, 0})
	// c.drawShadedTriangle(point{-200, -250, 0.3}, point{200, 50, 0.1}, point{20, 250, 1.0}, color{0, 255, 0})
	// c.drawShadedTriangle(point{-150, -50, 0.5}, point{-10, 50, 0.1}, point{20, 200, 0.9}, color{255, 255, 0})
	// c.drawShadedTriangle(point{0, 0, 0.9}, point{0, 130, 0.1}, point{20, 299, 0.9}, color{255, 0, 0})
	// render([]canvas{c})

	vA := vertex{-2, -0.5, 5}
	vB := vertex{-2, 0.5, 5}
	vC := vertex{-1, 0.5, 5}
	vD := vertex{-1, -0.5, 5}

	vAb := vertex{-2, -0.5, 6}
	vBb := vertex{-2, 0.5, 6}
	vCb := vertex{-1, 0.5, 6}
	vDb := vertex{-1, -0.5, 6}

	red := color{255, 0, 0}
	green := color{0, 255, 0}
	blue := color{0, 0, 255}

	c.drawLine(c.projectVertex(vA), c.projectVertex(vB), blue)
	c.drawLine(c.projectVertex(vB), c.projectVertex(vC), blue)
	c.drawLine(c.projectVertex(vC), c.projectVertex(vD), blue)
	c.drawLine(c.projectVertex(vD), c.projectVertex(vA), blue)

	c.drawLine(c.projectVertex(vAb), c.projectVertex(vBb), red)
	c.drawLine(c.projectVertex(vBb), c.projectVertex(vCb), red)
	c.drawLine(c.projectVertex(vCb), c.projectVertex(vDb), red)
	c.drawLine(c.projectVertex(vDb), c.projectVertex(vAb), red)

	c.drawLine(c.projectVertex(vA), c.projectVertex(vAb), green)
	c.drawLine(c.projectVertex(vB), c.projectVertex(vBb), green)
	c.drawLine(c.projectVertex(vC), c.projectVertex(vCb), green)
	c.drawLine(c.projectVertex(vD), c.projectVertex(vDb), green)
	render([]canvas{c})

}
