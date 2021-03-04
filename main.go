package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
)

const (
	defaultWidth, defaultHeight = 600, 320    // canvas size in pixels
	cells                       = 100         // number of grid cells
	xyrange                     = 30.0        // axis ranges (-xyrange..+xyrange)
	angle                       = math.Pi / 6 // angle of x, y axes (=30°)
)

var sin30, cos30 = math.Sin(angle), math.Cos(angle) // sin(30°), cos(30°)
var xyscale float64
var zscale float64
var width uint64
var height uint64

func main() {
	http.HandleFunc("/surface", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		r.ParseForm()
		var err error

		if qv, ok := r.Form["width"]; ok {
			width, err = strconv.ParseUint(qv[0], 10, 64)

			if err != nil {
				width = defaultWidth
			}
		} else {
			width = defaultWidth
		}

		if qv, ok := r.Form["height"]; ok {
			height, err = strconv.ParseUint(qv[0], 10, 64)

			if err != nil {
				height = defaultHeight
			}
		} else {
			height = defaultHeight
		}

		xyscale = float64(width) / 2 / xyrange // pixels per x or y unit
		zscale = float64(height) * 0.4
		buildSurface(w)
	})

	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

func buildSurface(out io.Writer) {
	fmt.Fprintf(out, "<svg xmlns='http://www.w3.org/2000/svg' "+
		"style='stroke: grey; fill: white; stroke-width: 0.7' "+
		"width='%d' hight='%d'>", width, height)

	for i := 0; i < cells; i++ {
		for j := 0; j < cells; j++ {
			ax, ay, az := corner(i+1, j)
			bx, by, bz := corner(i, j)
			cx, cy, cz := corner(i, j+1)
			dx, dy, dz := corner(i+1, j+1)

			if isSomeNumberInvalid(az, bz, cz, dz) {
				continue
			}

			maxZ := maxValue(az, bz, cz, dz) * zscale
			var redValue int
			var blueValue int

			if maxZ < 0 {
				blueValue = int(math.Abs(maxZ) * 25)
			} else if maxZ > 0 {
				redValue = int(maxZ * 25)
			}

			fmt.Fprintf(out, "<polygon fill='rgb(%d, 0, %d)' points='%g,%g %g,%g %g,%g %g,%g' />\n", redValue, blueValue, ax, ay, bx, by, cx, cy, dx, dy)
		}
	}

	fmt.Fprintln(out, "</svg>")
}

func isSomeNumberInvalid(numbers ...float64) bool {
	for _, n := range numbers {
		if math.IsNaN(n) || math.IsInf(n, 1) || math.IsInf(n, 0) || math.IsInf(n, -1) {
			return true
		}
	}

	return false
}

func maxValue(values ...float64) float64 {
	max := values[0]

	for _, v := range values {
		if v > max {
			max = v
		}
	}

	return max
}

func corner(i, j int) (float64, float64, float64) {
	// Find point (x, y) at corner of cell (i, j)
	x := xyrange * (float64(i)/cells - 0.5)
	y := xyrange * (float64(j)/cells - 0.5)

	// Compute surface height z
	z := f(x, y)

	// Project (x, y, z) isometrically onto 2-D SVG canvas (sx, sy)
	sx := float64(width)/2 + (x-y)*cos30*xyscale
	sy := float64(height)/2 + (x+y)*sin30*xyscale - z*zscale

	return sx, sy, z
}

func f(x, y float64) float64 {
	r := math.Hypot(x, y) // distance from (0, 0)

	return math.Sin(r) / r
}
