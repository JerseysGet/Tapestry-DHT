package util

import (
	"log"
	"math/bits"
	"unsafe"
)

const RADIX = 4

var DIGIT_SHIFT = bits.TrailingZeros(RADIX)
var DIGIT_MASK = uint64((1 << DIGIT_SHIFT) - 1)

var DIGITS = int(int(8*unsafe.Sizeof(uint64(0))) / DIGIT_SHIFT)

func DigitToChar(d uint64) rune {
	if d >= RADIX {
		log.Panicf("digit out of range: %d", d)
	}
	if d <= 9 {
		return rune('0' + d)
	}
	return rune('a' + d - 10)
}

func HashToString(h uint64) string {
	result := make([]rune, 0, DIGITS)
	for i := 0; i < DIGITS; i++ {
		d := uint64(h & DIGIT_MASK)
		result = append(result, DigitToChar(d))
		h >>= DIGIT_SHIFT
	}
	return string(result)
}

func StringToHash(s string) uint64 {
	if len(s) != DIGITS {
		log.Panicf("Invalid string length: got %d, want %d", len(s), DIGITS)
	}

	var h uint64 = 0
	for i := len(s) - 1; i >= 0; i-- {
		c := s[i]
		var digit uint64

		if '0' <= c && c <= '9' {
			digit = uint64(c - '0')
		} else if 'a' <= c && c <= 'f' {
			digit = uint64(c - 'a' + 10)
		} else {
			log.Panicf("Invalid character: %c", c)
		}

		if digit >= RADIX {
			log.Panicf("Digit out of bounds: %d", digit)
		}

		h = (h << DIGIT_SHIFT) | uint64(digit)
	}
	return h
}

func GetDigit(h uint64, i int) uint64 {
	if i < 0 || i >= DIGITS {
		log.Panicf("Index out of range: %d", i)
	}
	return (h >> (i * DIGIT_SHIFT)) & DIGIT_MASK
}

func CommonPrefixLen(a, b uint64) int {
	ret := 0
	for i := 0; i < DIGITS; i++ {
		if GetDigit(a, i) == GetDigit(b, i) {
			ret++
		} else {
			break
		}
	}
	return ret
}

func Assert(condition bool, msg string) {
	if !condition {
		log.Panic("Assertion failed: " + msg)
	}
}

// func main() {
// 	log.Printf("RADIX= %d", RADIX)
// 	log.Printf("DIGIT_SHIFT= %d", DIGIT_SHIFT)
// 	log.Printf("DIGIT_MASK= %d", DIGIT_MASK)
// 	log.Printf("DIGITS= %d", DIGITS)

// 	h1 := uint64(0xFFFFFFFFFFFFFFFF)
// 	str := HashToString(h1)
// 	log.Printf("Original hash: %x", h1)
// 	log.Printf("String form: %s", str)

// 	h2 := StringToHash(str)
// 	log.Printf("Recovered hash: %x", h2)

// 	if h1 != h2 {
// 		log.Panicf("Mismatch! h1 = %x, h2 = %x", h1, h2)
// 	} else {
// 		log.Println("Round-trip success! ✅")
// 	}

// 	log.Println("Testing GetDigit:")
// 	for i := 0; i < DIGITS; i++ {
// 		d1 := GetDigit(h1, i)
// 		d2 := GetDigit(h2, i)
// 		log.Printf("Digit %2d: %d vs %d", i, d1, d2)
// 		if d1 != d2 {
// 			log.Panicf("Mismatch at digit %d: %d != %d", i, d1, d2)
// 		}
// 	}
// 	log.Println("GetDigit works correctly! ✅")

// 	h3 := h1
// 	h4 := h1 ^ (1 << (2 * DIGIT_SHIFT))
// 	log.Printf("h1 = %v", HashToString(h1))
// 	log.Printf("h3 = %v (identical)", HashToString(h3))
// 	log.Printf("h4 = %v (differs at digit 2)", HashToString(h4))

// 	log.Printf("commonPrefixLen(h1, h3) = %d (expected %d)", commonPrefixLen(h1, h3), DIGITS)
// 	log.Printf("commonPrefixLen(h1, h4) = %d (expected 2)", commonPrefixLen(h1, h4))
// }
