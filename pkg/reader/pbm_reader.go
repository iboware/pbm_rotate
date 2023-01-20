package reader

import (
	"bufio"
	"errors"
	"fmt"
	"unicode"
)

type PBMReader struct {
	*bufio.Reader
	err error
}

type PBMHeader struct {
	Width    int
	Height   int
	Comments []string
}

func NewPBMReader(r *bufio.Reader) *PBMReader {
	return &PBMReader{
		Reader: r,
	}
}

func (nr PBMReader) Err() error {
	return nr.err
}

func (pr *PBMReader) GetNextByteAsRune() rune {
	if pr.err != nil {
		return 0
	}
	var b byte
	b, pr.err = pr.ReadByte()
	if pr.err != nil {
		return 0
	}
	return rune(b)
}

func (nr *PBMReader) GetHeader() (PBMHeader, bool) {
	var header PBMHeader

	rune1 := nr.GetNextByteAsRune()
	if rune1 != 'P' {
		return PBMHeader{}, false
	}
	rune2 := nr.GetNextByteAsRune()
	if rune2 != '1' {
		return PBMHeader{}, false
	}
	if !unicode.IsSpace(nr.GetNextByteAsRune()) {
		return PBMHeader{}, false
	}

	numbers, comments, err := nr.parseHeader()
	if err != nil {
		return PBMHeader{}, false
	}
	header.Comments = comments
	header.Width = numbers[0]
	header.Height = numbers[1]

	if nr.Err() != nil {
		return PBMHeader{}, false
	}

	return header, true
}

func (nr *PBMReader) parseHeader() ([]int, []string, error) {
	num := 0                         // Current number
	numbers := make([]int, 0, 2)     // All numbers
	cmt := make([]rune, 0, 100)      // Current comment
	comments := make([]string, 0, 1) // All comments

	const (
		ReadSpace = iota
		ReadNumber
		ReadComment
	)
	state := ReadSpace // Current state
	var prevState int  // Previous state

	var c rune
Loop:
	for {
		switch state {
		case ReadSpace:
			for c = nr.GetNextByteAsRune(); unicode.IsSpace(c); c = nr.GetNextByteAsRune() {
			}
			if c >= '0' && c <= '9' {
				state = ReadNumber
				num = int(c - '0')
			} else if c == '#' {
				state = ReadComment
				prevState = ReadSpace
			} else if c == 0 {
				return nil, nil, errors.New("unexpected eof in header")
			} else {
				return nil, nil, fmt.Errorf("unexpected char in header %q", c)
			}

		case ReadNumber:
			for c = nr.GetNextByteAsRune(); c >= '0' && c <= '9'; c = nr.GetNextByteAsRune() {
				num = num*10 + int(c-'0')
			}
			if unicode.IsSpace(c) {
				state = ReadSpace
				numbers = append(numbers, num)
				if len(numbers) == 2 {
					return numbers, comments, nil
				}
			} else if c == '#' {
				state = ReadComment
				prevState = ReadNumber
			} else if c == 0 {
				return nil, nil, errors.New("unexpected eof in header")
			} else {
				return nil, nil, fmt.Errorf("unexpected char in header %q", c)
			}

		case ReadComment:
			for c = nr.GetNextByteAsRune(); c != '\n' && c != '\r'; c = nr.GetNextByteAsRune() {
				cmt = append(cmt, c)
			}
			if len(cmt) > 0 && unicode.IsSpace(cmt[0]) {
				cmt = cmt[1:]
			}
			comments = append(comments, string(cmt))
			cmt = cmt[:0]
			state = prevState
			prevState = ReadComment

		default:
			break Loop
		}
	}
	return nil, nil, errors.New("unexpected eof in header")
}
