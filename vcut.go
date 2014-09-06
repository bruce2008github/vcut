// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cheggaaa/pb"

	"github.com/opennota/screengen"
)

var (
	n         = flag.Int("n", 100, "Number of screenshots")
	dirtmpl   = flag.String("d", "%n", "Output directory template (%n will be replaced with the video file name without extension)")
	fntmpl    = flag.String("f", "shot%03d.jpg", "Screenshot file name template")
	keepGoing = flag.Bool("keep-going", false, "Continue processing after an error")
)

func mkdir(name string) (string, error) {
	base := name
	for i := 0; ; i++ {
		_, err := os.Stat(name)
		if os.IsNotExist(err) {
			break
		}
		name = base + fmt.Sprintf("_%d", i)
	}
	return name, os.Mkdir(name, 0755)
}

func expand(tmpl, filename string) string {
	name := path.Base(filename)
	ext := path.Ext(name)
	name = name[:len(name)-len(ext)]
	return strings.Replace(tmpl, "%n", name, -1)
}

// writeImage writes image img to the file fn.
func writeImage(img image.Image, fn string) {
	f, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't create file: %v\n", err)
		if *keepGoing {
			return
		}
		os.Exit(1)
	}
	defer f.Close()

	err = jpeg.Encode(f, img, &jpeg.Options{Quality: 85})
	if err != nil {
		fmt.Fprintf(os.Stderr, "JPEG encoding error: %v\n", err)
		if *keepGoing {
			return
		}
		os.Exit(1)
	}
}

// GenerateScreenshots generates screenshots from the video file fn.
func GenerateScreenshots(fn string) {
	fmt.Println(fn)

	// Create a directory to store screenshots to.

	dname := expand(*dirtmpl, fn)
	dname, err := mkdir(dname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't create directory: %v\n", err)
		if *keepGoing {
			return
		}
		os.Exit(1)
	}

	// Generate screenshots.

	gen, err := screengen.NewGenerator(fn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading video file: %v\n", err)
		if *keepGoing {
			return
		}
		os.Exit(1)
	}
	defer gen.Close()

	inc := gen.Duration / int64(*n)

	bar := pb.StartNew(*n)
	var d int64
	for i := 0; i < *n; i++ {
		img, err := gen.Image(d)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Can't generate screenshot: %v\n", err)
			if *keepGoing {
				continue
			}
			os.Exit(1)
		}

		fn := filepath.Join(dname, fmt.Sprintf(*fntmpl, i))
		writeImage(img, fn)

		d += inc
		bar.Increment()
	}
	bar.Finish()
}

func main() {
	flag.Parse()
	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	for _, filename := range flag.Args() {
		GenerateScreenshots(filename)
	}
}
