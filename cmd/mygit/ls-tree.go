package main

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func lsTree() {
	if len(os.Args) < 4 || os.Args[2] != "--name-only" {
		fmt.Fprintf(os.Stderr, "usage: mygit ls-tree --name-only <treeSha>\n")
		os.Exit(1)
	}
	blob, err := readBlob(os.Args[3])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	nulIndex := bytes.IndexByte(blob, '\x00')
	if nulIndex == -1 {
		fmt.Println("Invalid tree data")
		return
	}

	// Validate the prefix
	prefixParts := strings.Fields(string(blob[:nulIndex]))
	if len(prefixParts) != 2 || prefixParts[0] != "tree" {
		fmt.Println("Invalid or missing 'tree' prefix")
		return
	}

	// Check if the length is a valid integer
	_, err = strconv.Atoi(prefixParts[1])
	if err != nil {
		fmt.Println("Invalid length in tree prefix")
		return
	}

	// Start parsing after the '\x00' of the prefix
	index := nulIndex + 1
	// Length of the SHA hash (20 bytes)
	shaLength := 20

	for index < len(blob) {
		// Find the index of the next null byte, marking the end of the name
		nulIndex := bytes.IndexByte(blob[index:], '\x00')
		if nulIndex == -1 {
			break
		}

		// Calculate the real null index relative to the entire array
		realNulIndex := index + nulIndex

		// The name starts right after the mode and space
		nameStart := index
		for nameStart < realNulIndex && blob[nameStart] != ' ' {
			nameStart++
		}
		nameStart++

		// Extract the name
		name := blob[nameStart:realNulIndex]
		fmt.Printf("%s\n", string(name))

		// Move the index to the start of the next entry (current position + name length + sha length)
		index = realNulIndex + 1 + shaLength
	}
}
