package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
)

// Usage: your_git.sh <command> <arg1> <arg2> ...
func main() {

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
			}
		}

		headFileContents := []byte("ref: refs/heads/main\n")
		if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
		}

		fmt.Println("Initialized git directory")
	case "cat-file": //./your_git.sh cat-file -p <blob_sha>
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
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}

func readBlob(sha1 string) ([]byte, error) {
	blobPath := fmt.Sprintf(".git/objects/%s/%s", sha1[:2], sha1[2:])
	// Open the compressed blob file
	compressedContent, err := os.Open(blobPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open blob: %s", err)
	}
	defer compressedContent.Close()

	// Decompress the blob using zlib
	r, err := zlib.NewReader(compressedContent)

	if err != nil {
		return nil, fmt.Errorf("failed to decompress blob: %w", err)
	}
	defer r.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return nil, fmt.Errorf("error reading blob: %w", err)
	}

	return buf.Bytes(), nil
}
