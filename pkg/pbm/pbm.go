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

	"github.com/iboware/pbm_rotate/pkg/reader"
)

type PBM struct {
	Header *reader.PBMHeader
	BitMap *[][]uint8
}

func DecodePBM(r io.Reader) (*PBM, error) {
	br := bufio.NewReader(r)
	pr := reader.NewPBMReader(br)

	// Parse the PBM header.
	header, ok := pr.GetHeader()
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
		Header: &header,
		BitMap: &bitmap,
	}, nil
}

func (pbm *PBM) RotateByAngle(angle float64) {
	bitmap := *pbm.BitMap
	cos := math.Cos(angle)
	sin := math.Sin(angle)

	newImage, shiftX, shiftY := emptyImageAndShift(bitmap, cos, sin)
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

	pbm.Header.Height = len(newImage)
	pbm.Header.Width = len(newImage[0])
	*pbm.BitMap = newImage
}

func emptyImageAndShift(bitmap [][]uint8, cos, sin float64) (img [][]uint8, shiftX, shiftY int) {
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
func (p *PBM) Save(w io.Writer) {
	wb := bufio.NewWriter(w)

	fmt.Fprintln(wb, "P1")
	for _, c := range p.Header.Comments {
		c = strings.Replace(c, "\n", " ", -1)
		c = strings.Replace(c, "\r", " ", -1)
		fmt.Fprintf(wb, "# %s\n", c)
	}

	fmt.Fprintf(wb, "%d %d\n", p.Header.Width, p.Header.Height)

	for _, row := range *p.BitMap {
		for _, col := range row {
			fmt.Fprintf(wb, "%d ", col)
		}
		fmt.Fprint(wb, "\n")
	}
	wb.Flush()
}
