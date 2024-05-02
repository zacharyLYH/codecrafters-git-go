package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func writeTree(dir string) {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit write-tree\n")
		os.Exit(1)
	}
	hash, err := dfs(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error at writeTree: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%x", hash)
}

func dfs(dir string) ([]byte, error) {
	// Open the directory
	directory, err := os.Open(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to open directory: %s", err)
	}
	defer directory.Close()

	fileInfos, err := directory.Readdir(-1)
	if err != nil {
		return nil, err
	}

	// Prepare a slice of entries that include the name for sorting
	type entry struct {
		name string
		mode string
		hash []byte
	}
	var entries []entry
	for _, fileInfo := range fileInfos {
		if fileInfo.Name() == ".git" {
			continue
		}
		fullPath := filepath.Join(dir, fileInfo.Name())
		mode := "100644" // Default mode for files
		if fileInfo.IsDir() {
			mode = "40000"
			hash, err := dfs(fullPath)
			if err != nil {
				return nil, err
			}
			entries = append(entries, entry{fileInfo.Name(), mode, hash})
		} else {
			fileData, err := os.ReadFile(fullPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read file: %s", err)
			}
			header := fmt.Sprintf("blob %d\x00", len(fileData))
			blobData := append([]byte(header), fileData...)
			hash, _, err := compress(blobData)
			if err != nil {
				return nil, fmt.Errorf("failed to compress file: %s", err)
			}
			entries = append(entries, entry{fileInfo.Name(), mode, hash})
		}
	}

	// Sort entries by name
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].name < entries[j].name
	})

	var data bytes.Buffer
	for _, e := range entries {
		data.WriteString(fmt.Sprintf("%s %s\x00%s", e.mode, e.name, e.hash))
	}

	treeData := fmt.Sprintf("tree %d\x00%s", data.Len(), data.Bytes())
	hash, compressedData, hashErr := compress([]byte(treeData))
	if hashErr != nil {
		return nil, fmt.Errorf("failed to compress treeObjData: %s", hashErr)
	}
	saveDirectory := fmt.Sprintf(".git/objects/%x/%x", hash[:1], hash[1:])
	if err := createNewFile(saveDirectory, compressedData); err != nil {
		return nil, fmt.Errorf("failed to write compressed data to %s: %v", saveDirectory, err)
	}
	return hash, nil
}
