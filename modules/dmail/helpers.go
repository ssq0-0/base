package dmail

import (
	cryptorand "crypto/rand"
	"errors"
	"math/big"
)

func generateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	for i := range result {
		index, err := generateRandomInt(0, len(charset)-1)
		if err != nil {
			return "", err
		}
		result[i] = charset[index]
	}

	return string(result), nil
}

func generateRandomInt(min, max int) (int, error) {
	if min > max {
		return 0, errors.New("min не может быть больше max")
	}

	diff := max - min + 1
	randomBig, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(diff)))
	if err != nil {
		return 0, err
	}
	return int(randomBig.Int64()) + min, nil
}
