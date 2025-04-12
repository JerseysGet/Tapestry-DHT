package main

import (
	"log"
	"math/bits"
	"unsafe"
)

const RADIX = 4
var DIGIT_SHIFT = bits.TrailingZeros(RADIX)
var DIGIT_MASK = (1 << DIGIT_SHIFT) - 1;
const DIGITS = 8 * unsafe.Sizeof(hash_t(0)) / RADIX

func main() {
	log.Printf("RADIX= %d", RADIX)	
	log.Printf("DIGIT_SHIFT= %d", DIGIT_SHIFT)	
	log.Printf("DIGIT_MASK= %d", DIGIT_MASK)	
	log.Printf("DIGITS= %d", DIGITS)	
	
}