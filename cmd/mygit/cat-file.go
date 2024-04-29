package main

import (
	"bytes"
	"fmt"
	"os"
)

func catFile() {
	if len(os.Args) < 4 || os.Args[2] != "-p" {
		fmt.Fprintf(os.Stderr, "usage: mygit git-cat -p <hash>\n")
		os.Exit(1)
	}
	blobSha := os.Args[3]
	blobContent, err := readBlob(blobSha)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read the blob: %s", blobContent)
		os.Exit(1)
	}
	// Locate the position of the null byte separating the header from the content
	nullByteIndex := bytes.IndexByte(blobContent, 0)
	if nullByteIndex == -1 {
		fmt.Fprintf(os.Stderr, "Invalid object format, missing null byte separator\n")
		os.Exit(1)
	}
	// Only print content after the null byte, without adding newline
	os.Stdout.Write(blobContent[nullByteIndex+1:])
}
