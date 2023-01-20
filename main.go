package main

import (
	"flag"
	"fmt"
	"math"
	"os"

	"github.com/iboware/pbm_rotate/pkg/pbm"
)

var usage = `Usage: pbm_rotate [options...]
Examples:
  # 
	pbm_rotate --file foo.pbm --angle 90
Options:
 -f --file	Path to a PBM file (only Magic Number P1 is supported).
 -a --angle	Angle of rotation (in degrees 90,-90 etc.)
`
var (
	filePath string
	angle    float64
)

// main.go
func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprint(usage))
	}
	flag.StringVar(&filePath, "file", "", "")
	flag.StringVar(&filePath, "f", "", "")
	flag.Float64Var(&angle, "angle", 0, "")
	flag.Float64Var(&angle, "a", 0, "")

	flag.Parse()
	if flag.NFlag() < 2 {
		usageAndExit("")
	}

	f, err := os.OpenFile(filePath, os.O_RDWR, os.ModeAppend)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	pbmImage, err := pbm.DecodePBM(f)
	if err != nil {
		panic(err)
	}

	radians := -1 * angle * (math.Pi / 180) // multiplied by -1 to make it clockwise.
	pbmImage.RotateByAngle(radians)

	f.Truncate(0)
	f.Seek(0, 0)
	pbmImage.Save(f)
}

func usageAndExit(msg string) {
	if msg != "" {
		fmt.Fprint(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n")
	}

	flag.Usage()
	os.Exit(0)
}
