package pbm

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"testing"
)

const PBMRaw = `P1
# test.pbm
7 5
0 0 0 0 0 0 0 
0 0 1 1 1 0 0 
0 0 0 1 0 0 0 
0 0 0 1 0 0 0 
0 0 0 0 0 0 0 
`

const PBMRaw90Rotated = `P1
# test.pbm
5 7
0 0 0 0 0 
0 0 0 0 0 
0 1 0 0 0 
0 1 1 1 0 
0 1 0 0 0 
0 0 0 0 0 
0 0 0 0 0 
`

func TestDecode(t *testing.T) {
	r := strings.NewReader(PBMRaw)

	pbmImage, err := Decode(r)
	if err != nil {
		t.Fatal(err)
	}

	if pbmImage.BitMap == nil {
		t.Fatal("bitmap does not exist")
	}
	if pbmImage.Config == nil {
		t.Fatal("config does not exist")
	}

	if len(pbmImage.Config.Comments) != 1 {
		t.Fatal("mismatching comments size")
	}

	if pbmImage.Config.Height != len(*pbmImage.BitMap) {
		t.Fatal("mismatching height and bitmap size")
	}

	if pbmImage.Config.Width != len((*pbmImage.BitMap)[0]) {
		t.Fatal("mismatching width and bitmap size")
	}

	if pbmImage.Config.Height != 5 {
		t.Fatalf("expected 5 as height but got %d", pbmImage.Config.Height)
	}

	if pbmImage.Config.Width != 7 {
		t.Fatalf("expected 7 as width but got %d", pbmImage.Config.Width)
	}
}

func TestEncode(t *testing.T) {
	var w bytes.Buffer

	r := strings.NewReader(PBMRaw)

	pbmImage, err := Decode(r)
	if err != nil {
		t.Fatal(err)
	}

	if err := pbmImage.Encode(&w); err != nil {
		t.Fatal(err)
	}
}

func TestRepeatDecodeEncode(t *testing.T) {
	var w bytes.Buffer

	r := strings.NewReader(PBMRaw)

	pbmImage, err := Decode(r)
	if err != nil {
		t.Fatal(err)
	}

	if err := pbmImage.Encode(&w); err != nil {
		t.Fatal(err)
	}
	firstImage := w.String()

	pbmImage, err = Decode(&w)
	if err != nil {
		t.Fatal(err)
	}

	if err := pbmImage.Encode(&w); err != nil {
		t.Fatal(err)
	}

	secondImage := w.String()
	if firstImage != secondImage {
		t.Fatal("encoding file has changed the image")
	}
}

func TestRotateByAngle(t *testing.T) {

	r := strings.NewReader(PBMRaw)

	pbmImage, err := Decode(r)
	if err != nil {
		t.Fatal(err)
	}

	radians := 90 * (math.Pi / 180)
	if err := pbmImage.RotateByAngle(radians); err != nil {
		t.Fatal(err)
	}

	if pbmImage.Config.Height != 7 {
		t.Fatalf("expected 7 as height but got %d", pbmImage.Config.Height)
	}

	if pbmImage.Config.Width != 5 {
		t.Fatalf("expected 5 as width but got %d", pbmImage.Config.Width)
	}

	var w bytes.Buffer
	if err = pbmImage.Encode(&w); err != nil {
		t.Fatal(err)
	}

	if PBMRaw90Rotated != w.String() {
		fmt.Println(w.String())
		fmt.Print(PBMRaw90Rotated)
		t.Fatal("expected output does not match to the rotated image")
	}
}
