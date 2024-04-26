package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
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
	case "hash-object":
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
		lengthStr := strconv.Itoa(len(fileData)) // Convert the length of the fileData to a string
		// Create the Git blob header with a null character
		header := []byte(fmt.Sprintf("blob %s\x00", lengthStr))
		// Concatenate the header and the original file data
		fileData = append(header, fileData...)
		hash, compressedData, compressErr := compress(fileData)
		if compressErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to compress the file: %s", compressedData)
			os.Exit(1)
		}
		saveDirectory := fmt.Sprintf(".git/objects/%s/%s", hash[:2], hash[2:])
		successWrite := createNewFile(saveDirectory, compressedData)
		if successWrite != nil {
			fmt.Fprintf(os.Stderr, "Failed to write compressed data to: %s. Tried to write %s. Error message: %s", saveDirectory, compressedData, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func createNewFile(filePath string, data []byte) error {
	// Ensure the directory exists
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, 0755) // Create the directory if it doesn't exist
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err) // Error creating the directory
	}
	// Open the file for writing, create it if it doesn't exist, truncate it if it does
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("failed to create file: %w", err) // Wrap the error with context
	}
	defer file.Close() // Ensure the file is closed when the function exits
	// Write the data to the file
	_, err = file.Write(data)
	if err != nil {
		fmt.Println("failed to write data to file")
		return fmt.Errorf("failed to write data to file: %w", err) // Handle write errors
	}
	return nil // Return nil to indicate success
}

func readFileReturnString(fileName string) ([]byte, error) {
	content, err := os.Open(fileName)
	check(err)
	defer content.Close()
	readBuffer := make([]byte, 4096)
	_, err = content.Read(readBuffer)
	check(err)
	return readBuffer, nil
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

func compress(data []byte) (string, []byte, error) {
	// Calculate the SHA-1 hash of the input data
	sha := sha1.New()
	sha.Write(data)
	hash := fmt.Sprintf("%x", sha.Sum(nil)) // Convert the hash to a hexadecimal string

	// Create a buffer for compressed data
	var buffer bytes.Buffer
	writer := zlib.NewWriter(&buffer)

	// Compress the data
	_, err := writer.Write(data) // Write data to the zlib writer
	if err != nil {
		return "", nil, err
	}
	writer.Close() // Finalize compression

	// Return the SHA-1 hash and the compressed data
	return hash, buffer.Bytes(), nil
}
