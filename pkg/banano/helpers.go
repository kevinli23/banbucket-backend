package banano

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/blake2b"
)

func arrayCrop(arr []uint8) []uint8 {
	length := len(arr) - 1
	cropped := make([]uint8, length)

	for i := 0; i < length; i++ {
		cropped[i] = arr[i+1]
	}

	return cropped
}

func uint4ToUint8(uint4 []uint8) []uint8 {
	length := len(uint4) / 2
	uint8Array := make([]uint8, length)

	for i := 0; i < length; i++ {
		uint8Array[i] = uint8(uint4[i*2]*16) + uint4[i*2+1]
	}

	return uint8Array
}

func uint5ToUint4(uint5 []uint8) []uint8 {
	length := (len(uint5) / 4) * 5

	uint4 := make([]uint8, length)

	for i := 1; i <= length; i++ {
		n := i - 1
		m := i % 5
		z := n - (i-m)/5
		var right uint8
		if z == 0 {
			right = 0
		} else {
			right = uint5[z-1] << (5 - m)
		}
		left := uint5[z] >> m

		uint4[n] = uint8((left + right) % 16)
	}

	return uint4
}

func stringToUint5(crop string) []uint8 {
	length := len(crop)
	stringArray := strings.Split(crop, "")

	uint5 := make([]uint8, length)

	for i := 0; i < length; i++ {
		uint5[i] = uint8(strings.Index("13456789abcdefghijkmnopqrstuwxyz", stringArray[i]))
	}

	return uint5
}

func getBlake2BHash(size int, input []uint8) ([]uint8, error) {
	ctx, err := blake2b.New(5, nil)
	if err != nil {
		return nil, err
	}
	ctx.Write(input)
	return ctx.Sum(nil), nil
}

func reverseuint8(input []uint8) []uint8 {
	for i := 0; i < len(input)/2; i++ {
		j := len(input) - i - 1
		input[i], input[j] = input[j], input[i]
	}

	return input
}

func uint8ToUint4(uint8Arr []uint8) []uint8 {

	uint4 := make([]uint8, len(uint8Arr)*2)

	for i := 0; i < len(uint8Arr); i++ {
		uint4[i*2] = (uint8Arr[i] / 16)
		uint4[i*2+1] = uint8Arr[i] % 16
	}

	return uint4
}

func uint4ToHex(uint4 []uint8) string {
	var hex string
	for i := 0; i < len(uint4); i++ {
		hex += strings.ToUpper(fmt.Sprintf("%x", uint4[i]))
	}
	return hex
}
