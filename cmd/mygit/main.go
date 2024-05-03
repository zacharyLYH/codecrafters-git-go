package main

import (
	"fmt"
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
		catFile()
	case "hash-object":
		hashObject()
	case "ls-tree":
		lsTree()
	case "write-tree":
		cwd, _ := os.Getwd()
		writeTree(cwd)
	case "commit-tree":
		// tree_sha := os.Args[2]
		// tree, _ := readBlob(tree_sha)
		// fmt.Println(string(tree))
		// commitSha := os.Args[4]
		// commit, _ := readBlob(commitSha)
		// fmt.Println(string(commit))
		commitTree()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
