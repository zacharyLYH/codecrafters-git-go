package main

import (
	"fmt"
	"os"
)

func hashObject() {
	if len(os.Args) < 4 || os.Args[2] != "-w" {
		fmt.Fprintf(os.Stderr, "usage: mygit hash-object -w <fileName>\n")
		os.Exit(1)
	}
	fileNameToRead := os.Args[3]
	fileData, err := readFileReturnString(fileNameToRead)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read the file: %s", fileData)
		os.Exit(1)
	}
	// Create the Git blob header with a null character
	header := fmt.Sprintf("blob %d\x00", len(fileData))
	// Concatenate the header and the original file data
	fileData = append([]byte(header), fileData...)
	hash, compressedData, compressErr := compress(fileData)
	if compressErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to compress the file: %s", compressedData)
		os.Exit(1)
	}
	saveDirectory := fmt.Sprintf(".git/objects/%x/%x", hash[:1], hash[1:])
	successWrite := createNewFile(saveDirectory, compressedData)
	if successWrite != nil {
		fmt.Fprintf(os.Stderr, "Failed to write compressed data to: %s. Tried to write %s. Error message: %s", saveDirectory, compressedData, err)
		os.Exit(1)
	}
	fmt.Printf("%x\n", hash)
}
