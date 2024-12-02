package utils

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

func ReadAbi(relativePath string) (*abi.ABI, error) {
	rootDir := GetRootDir()
	absolutePath := filepath.Join(rootDir, relativePath)

	file, err := os.ReadFile(absolutePath)
	if err != nil {
		log.Printf("failed to read abi file: %v, path: %s", err, absolutePath)
		return nil, err
	}

	abi, err := abi.JSON(strings.NewReader(string(file)))
	if err != nil {
		log.Printf("failed decode abi: %v", err)
		return nil, err
	}

	return &abi, nil
}

func GetRootDir() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}
