package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

var (
	version = "0.0.1"
	info    *log.Logger
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	args := os.Args
	processArguments(args)
}

func processArguments(args []string) {
	if len(args) == 1 {
		printHelp()
	}
	if len(args) == 2 {
		info = log.New(ioutil.Discard, "", 0)
		evaluate(args[1])
	}
	if len(args) == 3 {
		if args[2] == "-v" {
			info = log.New(os.Stdout, "", 0)
			evaluate(args[1])
		} else {
			printHelp()
		}
	}
}

func evaluate(filePath string) {
	info.Println(versionInfo())

	checkFileExists(filePath)

	contents := getContents(filePath)

	lines := bytes.Split(contents, []byte("\n"))

	lines = removeComments(lines)

	for _, i := range lines {
		tokens := bytes.Split(i, []byte(" "))
		// Interpret and verify syntax of line

		// Insert interpreted instruction to vm

	}
	// If there were no errors:
	// Loop interpreted instruction set
	// Execute each command in sequential order
	// Immediately stop if a command execution results in an error
}

func removeComments(lines [][]byte) [][]byte {
	var syntaxLines [][]byte
	for _, i := range lines {
		if len(i) > 0 {
			if i[0] != []byte("#")[0] {
				syntaxLines = append(syntaxLines, i)
			}
		}
	}
	return syntaxLines
}

func getContents(filePath string) []byte {
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal("Unable to read file:", filePath)
	}
	info.Println("Loaded contents of", filePath)
	return contents
}

func checkFileExists(filePath string) {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatal("File not found:", filePath)
		} else {
			log.Println("Unexpected error for file:", filePath)
			log.Fatal(err)
		}
	}
}

func printHelp() {
	fmt.Println("Gosh Usage:")
	fmt.Println("\tgosh script_file [-v]")
}

func versionInfo() string {
	return ("GOSH INTERPRETER VER " + version)
}
