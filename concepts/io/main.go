package main

import (
	"fmt"
	"io"
	"strings"
)

func countLetter(r io.Reader) (map[string]int, error) {
	buf := make([]byte, 2048)
	out := map[string]int{}
	for {
		n, err := r.Read(buf)
		for _, b := range buf[:n] {
			if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') {
				out[string(b)]++
			}
		}
		if err == io.EOF {
			return out, nil
		}
		if err != nil {
			return nil, err
		}
	}

}

type ourReader struct{}

func (*ourReader) Read(p []byte) (n int, er error) {
	var out int
	var err error

	return out, err
}

func main() {
	sr := strings.NewReader("io is quite fun, I say!")
	buf := make([]byte, 100)

	read, _ := sr.Read(buf)
	fmt.Printf(" %v bytes written to buf:\n%v\n", read, buf)

	for _, i := range buf {
		fmt.Printf(string(i))
	}
}
