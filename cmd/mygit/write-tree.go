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
	// Read all files and directories in the current directory
	fileInfos, err := directory.Readdir(-1)
	if err != nil {
		return nil, err
	}
	// Process each file/directory
	var entries []string
	for _, fileInfo := range fileInfos {
		fullPath := filepath.Join(dir, fileInfo.Name())
		if fileInfo.IsDir() { //040000
			if fileInfo.Name() == ".git" {
				continue
			}
			if hash, err := dfs(fullPath); err != nil {
				return nil, err
			} else {
				entries = append(entries, fmt.Sprintf("%s %s\x00%s", "40000", fileInfo.Name(), hash))
			}
		} else {
			fileData, err := os.ReadFile(fullPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read the file: %s", fileData)
			}
			header := fmt.Sprintf("blob %d\x00", len(fileData))
			fileData = append([]byte(header), fileData...)
			hash, _, hashErr := compress(fileData)
			if hashErr != nil {
				return nil, fmt.Errorf("failed to compress file: %s", hashErr)
			}
			entries = append(entries, fmt.Sprintf("%s %s\x00%s", "100644", fileInfo.Name(), hash))
		}
	}
	// Sort entries as required by Git
	sort.Strings(entries)

	var data bytes.Buffer
	for _, entry := range entries {
		data.WriteString(entry)
	}

	treeData := fmt.Sprintf("tree %d\x00%s", data.Len(), data.Bytes())
	hash, compressedData, hashErr := compress([]byte(treeData))
	if hashErr != nil {
		return nil, fmt.Errorf("failed to compress treeObjData: %s", hashErr)
	}
	saveDirectory := fmt.Sprintf(".git/objects/%x/%x", hash[:1], hash[1:])
	successWrite := createNewFile(saveDirectory, compressedData)
	if successWrite != nil {
		fmt.Fprintf(os.Stderr, "Failed to write compressed data to: %s. Tried to write %s. Error message: %s", saveDirectory, compressedData, err)
		os.Exit(1)
	}
	return hash, nil
}
