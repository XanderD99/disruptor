package util

import (
	"crypto/rand"
	"math/big"
)

const floatPrecision = 1000000

func RandomInt(minimum, maximum int) int {
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(maximum+1-minimum)))
	if err != nil {
		panic(err) // Handle error appropriately in production code
	}
	n := nBig.Int64()
	return int(n) + minimum
}

func RandomFloat(minimum, maximum float64) float64 {
	minInt := int(minimum * floatPrecision)
	maxInt := int(maximum * floatPrecision)

	return float64(RandomInt(minInt, maxInt)) / floatPrecision
}
