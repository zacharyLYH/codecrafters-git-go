package main

import (
	"fmt"
	"os"
	"time"
)

func commitTree() {
	if len(os.Args) != 5 && len(os.Args) != 7 {
		fmt.Fprintf(os.Stderr, "usage: mygit commit-tree <treeHash> -m <message> OR mygit commit-tree <treeHash> -p <parentTreeHash> -m <message>\n")
		os.Exit(1)
	}
	if len(os.Args) == 5 {
		if os.Args[3] != "-m" {
			fmt.Fprintf(os.Stderr, "usage: mygit commit-tree <treeHash> -m <message>\n")
			os.Exit(1)
		}
	}
	if len(os.Args) == 7 {
		if os.Args[3] != "-p" && os.Args[5] != "-m" {
			fmt.Fprintf(os.Stderr, "usage: mygit commit-tree <treeHash> -p <parentTreeHash> -m <message>\n")
			os.Exit(1)
		}
	}
	treeHash := os.Args[2]
	parentHash := ""
	commitMessage := ""
	if len(os.Args) == 7 {
		parentHash = os.Args[4]
		commitMessage = os.Args[6]
	} else {
		commitMessage = os.Args[4]

	}
	fmt.Printf("%x\n", commit(treeHash, parentHash, commitMessage))
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func commit(treeHash string, parentHash string, commitMessage string) []byte {
	email := "xyz@gmail.com"
	name := "xyz"
	timestamp := time.Now().Unix()
	_, offset := time.Now().Zone()
	hours := abs(offset) / 3600
	minutes := (abs(offset) % 3600) / 60
	sign := '+'
	if offset < 0 {
		sign = '-'
	}
	authorTime := fmt.Sprintf("%d %c%02d%02d", timestamp, sign, hours, minutes)
	lines := fmt.Sprintf("tree %s\n", treeHash)
	if parentHash != "" {
		lines += fmt.Sprintf("parent %s\n", parentHash)
	}
	lines += fmt.Sprintf("author %s %s %v\n", name, email, authorTime)
	lines += fmt.Sprintf("committer %s %s %v\n\n", name, email, authorTime)
	lines += fmt.Sprintf("%s\n", commitMessage)

	header := fmt.Sprintf("commit %d\x00", len(lines))
	// Concatenate the header and the original file data
	fileData := append([]byte(header), lines...)
	hash, compressedData, compressErr := compress(fileData)
	if compressErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to compress the file: %s", compressedData)
		os.Exit(1)
	}
	saveDirectory := fmt.Sprintf(".git/objects/%x/%x", hash[:1], hash[1:])
	successWrite := createNewFile(saveDirectory, compressedData)
	if successWrite != nil {
		fmt.Fprintf(os.Stderr, "Failed to write compressed data to: %s. Tried to write %s. Error message: %s", saveDirectory, compressedData, successWrite)
		os.Exit(1)
	}
	return hash
}
