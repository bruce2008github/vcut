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
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cheggaaa/pb"
)

var (
	n       = flag.Int("n", 100, "Number of screenshots")
	dirtmpl = flag.String("d", "%n", "Output directory template (%n will be replaced with the video file name without extension)")
	fntmpl  = flag.String("f", "shot%03d.jpg", "Screenshot file name template")
)

var (
	rDuration = regexp.MustCompile(`Duration: (\d\d):(\d\d):(\d\d)`)
)

func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

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

func main() {
	flag.Parse()
	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	for _, filename := range flag.Args() {
		fmt.Println(filename)

		// Create a directory to store screenshots to.
		dname := expand(*dirtmpl, filename)
		dname, err := mkdir(dname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Can't create directory: %v\n", err)
			os.Exit(1)
		}

		// Get movie duration.
		cmd := exec.Command("ffmpeg", "-i", filename)
		out, err := cmd.CombinedOutput()
		match := rDuration.FindStringSubmatch(string(out))
		if match == nil {
			fmt.Fprintf(os.Stderr, "Can't get movie duration: %v:\nffmpeg's output:\n", err)
			fmt.Fprint(os.Stderr, string(out))
			os.Exit(1)
		}

		h := atoi(match[1])
		m := atoi(match[2])
		s := atoi(match[3])
		d := time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second
		inc := d / time.Duration(*n)

		// Make screenshots.
		bar := pb.StartNew(*n)
		d = 0
		for i := 0; i < *n; i++ {
			fname := fmt.Sprintf(*fntmpl, i)
			cmd := exec.Command("ffmpeg", "-ss", fmt.Sprintf("%02d:%02d:%02d", int(d.Hours()),
				int(d.Minutes())%60, int(d.Seconds())%60), "-i", filename, "-f",
				"image2", "-vframes", "1", filepath.Join(dname, fname))
			out, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Can't get frame: %v\nffmpeg's output:\n", err)
				fmt.Fprint(os.Stderr, string(out))
				os.Exit(1)
			}
			d += inc
			bar.Increment()
		}
		bar.Finish()
	}
}
