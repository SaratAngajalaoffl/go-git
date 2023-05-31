package main

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	objectsDir = ".go-git/objects"
	commitsDir = ".go-git/commits"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: mygit <command>")
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "init":
		initRepo()
	case "add":
		add(os.Args[2:])
	case "commit":
		commit(os.Args[2:])
	case "log":
		logCommand()
	case "push":
		logCommand()
	case "pull":
		logCommand()
	case "remote":
		logCommand()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func initRepo() {
	err := os.Mkdir(".go-git", 0755)
	if err != nil {
		log.Fatalf("Error creating .go-git directory: %v", err)
	}
	err = os.Mkdir(filepath.Join(".go-git", "objects"), 0755)
	if err != nil {
		log.Fatalf("Error creating objects directory: %v", err)
	}
	err = os.Mkdir(filepath.Join(".go-git", "commits"), 0755)
	if err != nil {
		log.Fatalf("Error creating commits directory: %v", err)
	}
	err = ioutil.WriteFile(filepath.Join(".go-git", "HEAD"), []byte("ref: refs/heads/master\n"), 0644)
	if err != nil {
		log.Fatalf("Error creating HEAD file: %v", err)
	}
	fmt.Println("Initialized empty Git repository in .go-git/")
}

func add(files []string) {
	if len(files) == 0 {
		fmt.Println("Nothing specified, nothing added.")
		return
	}

	for _, file := range files {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatalf("Error reading file %s: %v", file, err)
		}

		objectPath := getObjectPath(content)

		if _, err := os.Stat(objectPath); os.IsNotExist(err) {
			err := ioutil.WriteFile(objectPath, content, 0644)
			if err != nil {
				log.Fatalf("Error writing object file %s: %v", objectPath, err)
			}
		}
	}
}

func getObjectPath(content []byte) string {
	sha1 := sha1sum(content)
	return filepath.Join(objectsDir, sha1[:])
}

func sha1sum(content []byte) string {
	h := sha1.New()
	h.Write(content)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func commit(args []string) {
	if len(args) == 0 {
		fmt.Println("Please provide a commit message.")
		os.Exit(1)
	}

	message := strings.Join(args, " ")

	// Get the list of objects to include in the commit
	objects, err := ioutil.ReadDir(objectsDir)
	if err != nil {
		log.Fatalf("Error reading objects directory %s: %v", objectsDir, err)
	}

	// Create a new commit object
	var objectLines []string
	for _, object := range objects {
		content, err := ioutil.ReadFile(filepath.Join(objectsDir, object.Name()))
		if err != nil {
			log.Fatalf("Error reading object file %s: %v", object.Name(), err)
		}
		objectLines = append(objectLines, object.Name()+" "+sha1sum(content))
	}
	commitContent := []byte(fmt.Sprintf("tree %s\n\n%s\n", sha1sum([]byte(strings.Join(objectLines, "\n"))), message))
	commitPath := filepath.Join(commitsDir, time.Now().Format("2006-01-02T15-04-05"))
	if err := ioutil.WriteFile(commitPath, commitContent, 0644); err != nil {
		log.Fatalf("Error writing commit file %s: %v", commitPath, err)
	}
	fmt.Printf("Created commit %s\n", commitPath)
}

func logCommand() {
	commits, err := ioutil.ReadDir(commitsDir)
	if err != nil {
		log.Fatalf("Error reading commits directory %s: %v", commitsDir, err)
	}
	for i := len(commits) - 1; i >= 0; i-- {
		commit := commits[i]
		content, err := ioutil.ReadFile(filepath.Join(commitsDir, commit.Name()))
		if err != nil {
			log.Fatalf("Error reading commit file %s: %v", commit.Name(), err)
		}
		fmt.Printf("\n%s\n\n", strings.TrimSpace(string(content)))
	}
}
