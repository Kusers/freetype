// Copyright 2010 The Freetype-Go Authors. All rights reserved.
// Use of this source code is governed by your choice of either the
// FreeType License or the GNU General Public License version 2 (or
// any later version), both of which can be found in the LICENSE file.

// This program visualizes the quadratic approximation to the circle, used to
// implement round joins when stroking paths. The approximation is used in the
// stroking code for arcs between 0 and 45 degrees, but is visualized here
// between 0 and 90 degrees. The discrepancy between the approximation and the
// true circle is clearly visible at angles above 65 degrees.
package main

import (
	"bufio"
	"fmt"
	"image"
	"image/png"
	"log"
	"math"
	"os"

	"freetype-go.googlecode.com/hg/freetype/raster"
)

func main() {
	const (
		n = 17
		r = 256 * 80
	)
	s := raster.Fix32(r * math.Sqrt(2) / 2)
	t := raster.Fix32(r * math.Tan(math.Pi/8))

	m := image.NewRGBA(800, 600)
	for y := 0; y < m.Height(); y++ {
		for x := 0; x < m.Width(); x++ {
			m.Pixel[y][x] = image.RGBAColor{63, 63, 63, 255}
		}
	}
	mp := raster.NewRGBAPainter(m)
	mp.SetColor(image.Black)
	z := raster.NewRasterizer(800, 600)

	for i := 0; i < n; i++ {
		cx := raster.Fix32(25600 + 51200*(i%4))
		cy := raster.Fix32(2560 + 32000*(i/4))
		c := raster.Point{cx, cy}
		theta := math.Pi * (0.5 + 0.5*float64(i)/(n-1))
		dx := raster.Fix32(r * math.Cos(theta))
		dy := raster.Fix32(r * math.Sin(theta))
		d := raster.Point{dx, dy}
		// Draw a quarter-circle approximated by two quadratic segments,
		// with each segment spanning 45 degrees.
		z.Start(c)
		z.Add1(c.Add(raster.Point{r, 0}))
		z.Add2(c.Add(raster.Point{r, t}), c.Add(raster.Point{s, s}))
		z.Add2(c.Add(raster.Point{t, r}), c.Add(raster.Point{0, r}))
		// Add another quadratic segment whose angle ranges between 0 and 90 degrees.
		// For an explanation of the magic constants 22, 150, 181 and 256, read the
		// comments in the freetype/raster package.
		dot := 256 * d.Dot(raster.Point{0, r}) / (r * r)
		multiple := raster.Fix32(150 - 22*(dot-181)/(256-181))
		z.Add2(c.Add(raster.Point{dx, r + dy}.Mul(multiple)), c.Add(d))
		// Close the curve.
		z.Add1(c)
	}
	z.Rasterize(mp)

	for i := 0; i < n; i++ {
		cx := raster.Fix32(25600 + 51200*(i%4))
		cy := raster.Fix32(2560 + 32000*(i/4))
		for j := 0; j < n; j++ {
			theta := math.Pi * float64(j) / (n - 1)
			dx := raster.Fix32(r * math.Cos(theta))
			dy := raster.Fix32(r * math.Sin(theta))
			m.Set(int((cx+dx)/256), int((cy+dy)/256), image.RGBAColor{255, 255, 0, 255})
		}
	}

	// Save that RGBA image to disk.
	f, err := os.Open("out.png", os.O_CREAT|os.O_WRONLY, 0600)
	if err != nil {
		log.Stderr(err)
		os.Exit(1)
	}
	defer f.Close()
	b := bufio.NewWriter(f)
	err = png.Encode(b, m)
	if err != nil {
		log.Stderr(err)
		os.Exit(1)
	}
	err = b.Flush()
	if err != nil {
		log.Stderr(err)
		os.Exit(1)
	}
	fmt.Println("Wrote out.png OK.")
}
