package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"

	pbm "github.com/iboware/pbm_rotate/pkg"
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
		log.Fatalf("Failed opening the image: %s", err.Error())
	}
	defer f.Close()

	pbmImage, err := pbm.Decode(f)
	if err != nil {
		log.Fatalf("Failed decoding the image: %s", err.Error())
	}

	radians := -1 * angle * (math.Pi / 180) // multiplied by -1 to make it clockwise.
	pbmImage.RotateByAngle(radians)

	f.Truncate(0)
	f.Seek(0, 0)
	if err := pbmImage.Encode(f); err != nil {
		log.Fatalf("Failed encoding the image: %s", err.Error())
	}
}

func usageAndExit(msg string) {
	if msg != "" {
		fmt.Fprint(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n")
	}

	flag.Usage()
	os.Exit(0)
}
