package pbm

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

func NewPBMReader(r *bufio.Reader) *PBMReader {
	return &PBMReader{
		Reader: r,
	}
}

// Err returns sticky error of the reader.
func (nr PBMReader) Err() error {
	return nr.err
}

// GetNextByteAsRune reads a byte as Rune.
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

// GetConfig reads header section of the PBM P1 and returns the PBM header configuration
func (nr *PBMReader) GetConfig() (Config, bool) {
	var config Config

	rune1 := nr.GetNextByteAsRune()
	if rune1 != 'P' {
		return Config{}, false
	}
	rune2 := nr.GetNextByteAsRune()
	if rune2 != '1' {
		return Config{}, false
	}
	if !unicode.IsSpace(nr.GetNextByteAsRune()) {
		return Config{}, false
	}

	numbers, comments := nr.parseHeader()
	if nr.Err() != nil {
		return Config{}, false
	}
	config.Comments = comments
	config.Width = numbers[0]
	config.Height = numbers[1]

	if nr.Err() != nil {
		return Config{}, false
	}

	return config, true
}

func (nr *PBMReader) parseHeader() ([]int, []string) {
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
				nr.err = errors.New("unexpected eof in header")
				return nil, nil
			} else {
				nr.err = fmt.Errorf("unexpected char in header %q", c)
				return nil, nil
			}

		case ReadNumber:
			for c = nr.GetNextByteAsRune(); c >= '0' && c <= '9'; c = nr.GetNextByteAsRune() {
				num = num*10 + int(c-'0')
			}
			if unicode.IsSpace(c) {
				state = ReadSpace
				numbers = append(numbers, num)
				if len(numbers) == 2 {
					return numbers, comments
				}
			} else if c == '#' {
				state = ReadComment
				prevState = ReadNumber
			} else if c == 0 {
				nr.err = errors.New("unexpected eof in header")
				return nil, nil
			} else {
				nr.err = fmt.Errorf("unexpected char in header %q", c)
				return nil, nil
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
	nr.err = errors.New("unexpected eof in header")
	return nil, nil
}
