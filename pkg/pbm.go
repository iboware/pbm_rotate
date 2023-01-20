package pbm

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"
	"unicode"
)

type PBM struct {
	Config *Config
	BitMap *[][]uint8
}

type Config struct {
	Width    int      // Width of the image
	Height   int      // Height of the image
	Comments []string // Comments of the image
}

// Decode reads PBM file from an io.Reader and returns it as a PBM object
func Decode(r io.Reader) (*PBM, error) {
	br := bufio.NewReader(r)
	pr := NewPBMReader(br)

	// Parse the PBM header.
	header, ok := pr.GetConfig()
	if !ok {
		err := pr.Err()
		if err == nil {
			err = errors.New("invalid header")
		}
		return nil, err
	}
	errorFunc := func() (*PBM, error) {
		err := pr.Err()
		if err == nil {
			err = errors.New("failed to parse PBM data")
		}
		return nil, err
	}

	bitmap := make([][]uint8, header.Height)
	bitmap[0] = make([]uint8, header.Width)

	for col, row := 0, 0; col*row < (header.Width-1)*(header.Height-1); {
		ch := pr.GetNextByteAsRune()

		switch {
		case pr.Err() != nil:
			return errorFunc()
		case unicode.IsSpace(ch):
			continue
		case ch == '0' || ch == '1':
			bitmap[row][col] = uint8(ch - '0')
			if col == header.Width-1 && row < header.Height-1 {
				row++
				bitmap[row] = make([]uint8, header.Width)
				col = -1
			}
			col++
		default:
			return errorFunc()
		}
	}

	return &PBM{
		Config: &header,
		BitMap: &bitmap,
	}, nil
}

// RotateByAngle rotates an image by angle using rotation matrix.
// https://en.wikipedia.org/wiki/Rotation_matrix
func (p *PBM) RotateByAngle(angle float64) error {
	if p == nil {
		return errors.New("no image loaded")
	}

	if p.BitMap == nil {
		return errors.New("no bitmap loaded")
	}

	if p.Config == nil {
		return errors.New("no header loaded")
	}

	bitmap := *p.BitMap
	cos := math.Cos(angle)
	sin := math.Sin(angle)

	newImage, shiftX, shiftY := resizeAndShiftImage(bitmap, cos, sin)
	var wg sync.WaitGroup
	for i := 0; i < len(bitmap); i++ {
		for j := 0; j < len(bitmap[i]); j++ {
			wg.Add(1)
			go func(i, j int) {
				xNew := int(float64(i)*cos-float64(j)*sin) + shiftX
				yNew := int(float64(i)*sin+float64(j)*cos) + shiftY
				newImage[xNew][yNew] = bitmap[i][j]
				wg.Done()
			}(i, j)
		}
	}
	wg.Wait()

	p.Config.Height = len(newImage)
	p.Config.Width = len(newImage[0])
	*p.BitMap = newImage

	return nil
}

// Encode writes the PBM image into disk by given io.Writer
func (p *PBM) Encode(w io.Writer) error {
	if p == nil {
		return errors.New("no image loaded")
	}

	if p.BitMap == nil {
		return errors.New("no bitmap loaded")
	}

	if p.Config == nil {
		return errors.New("no header loaded")
	}

	wb := bufio.NewWriter(w)

	fmt.Fprintln(wb, "P1")
	for _, c := range p.Config.Comments {
		c = strings.Replace(c, "\n", " ", -1)
		c = strings.Replace(c, "\r", " ", -1)
		fmt.Fprintf(wb, "# %s\n", c)
	}

	fmt.Fprintf(wb, "%d %d\n", p.Config.Width, p.Config.Height)

	for _, row := range *p.BitMap {
		for _, col := range row {
			fmt.Fprintf(wb, "%d ", col)
		}
		fmt.Fprint(wb, "\n")
	}
	if err := wb.Flush(); err != nil {
		return err
	}

	return nil
}

// resizeAndShiftImage resizes and shifts bits in the matrix of bitmap.
func resizeAndShiftImage(bitmap [][]uint8, cos, sin float64) (img [][]uint8, shiftX, shiftY int) {
	var lock = sync.Mutex{}
	var wg sync.WaitGroup
	var xMax, yMax, xLow, yLow int
	for i := 0; i < len(bitmap); i++ {
		for j := 0; j < len(bitmap[i]); j++ {
			wg.Add(1)
			go func(i, j int) {
				defer lock.Unlock()
				defer wg.Done()
				// calculate the new x and y coordinates of the pixels
				xNew := int(float64(i)*cos - float64(j)*sin)
				yNew := int(float64(i)*sin + float64(j)*cos)
				lock.Lock()

				// set boundaries to the new values
				if xNew > xMax {
					xMax = xNew
				}
				if yNew > yMax {
					yMax = yNew
				}
				if xNew < xLow {
					xLow = xNew
				}
				if yNew < yLow {
					yLow = yNew
				}
			}(i, j)
		}
	}
	wg.Wait()
	shiftX = int(math.Abs(float64(xLow)))
	width := xMax + shiftX
	shiftY = int(math.Abs(float64(yLow)))
	height := yMax + shiftY
	newImage := make([][]uint8, width+1)
	for i := 0; i < len(newImage); i++ {
		newImage[i] = make([]uint8, height+1)
	}
	return newImage, shiftX, shiftY
}
