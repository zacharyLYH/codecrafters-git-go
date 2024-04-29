package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

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
	file_contents, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err)
	}
	return file_contents, nil
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

func compress(data []byte) ([]byte, []byte, error) {
	// Calculate the SHA-1 hash of the input data
	sha := sha1.New()
	sha.Write(data)
	hash := sha.Sum(nil) // Convert the hash to a hexadecimal string

	// Create a buffer for compressed data
	var buffer bytes.Buffer
	writer := zlib.NewWriter(&buffer)

	// Compress the data
	_, err := writer.Write(data) // Write data to the zlib writer
	if err != nil {
		return nil, nil, err
	}
	writer.Close() // Finalize compression

	// Return the SHA-1 hash and the compressed data
	return hash, buffer.Bytes(), nil
}
