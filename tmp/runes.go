// +build main

package main

import U "unicode/utf8"
import "os"

const h = 0x2550
const v = 0x2551

var box = []rune{
	0x2554, 0x2566, 0x2557,
	0x2560, 0x256C, 0x2563,
	0x255A, 0x2569, 0x255D,
}

func putr(r rune) {
	bb := make([]byte, 4)
	cc := U.EncodeRune(bb, r)
	os.Stdout.Write(bb[:cc])
}

func main() {
	putr(box[0])
	for j := 0; j < 10; j++ {
		putr(box[1])
		for k := 0; k < j; k++ {
			putr(h)
		}
	}
	putr(box[2])
	putr('\n')

	for i := 0; i < 5; i++ {
		putr(box[3])
		for j := 0; j < 10; j++ {
			putr(box[4])
			for k := 0; k < j; k++ {
				putr(h)
			}
		}
		putr(box[5])
		putr('\n')

		for g := 0; g < i; g++ {
			putr(v)
			for j := 0; j < 10; j++ {
				putr(v)
				for k := 0; k < j; k++ {
					putr(' ')
				}
			}
			putr(v)
			putr('\n')
		}
	}

	putr(box[6])
	for j := 0; j < 10; j++ {
		putr(box[7])
		for k := 0; k < j; k++ {
			putr(h)
		}
	}
	putr(box[8])
	putr('\n')
}
