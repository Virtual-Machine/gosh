package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
)

var (
	version = "0.0.1"
	info    *log.Logger
	method  = make(map[string]int)
	regex   *regexp.Regexp
)

func init() {
	log.SetFlags(0)
	method["create"] = 2
	method["cd"] = 2
	method["rm"] = 2
	method["cp"] = 3
	method["mv"] = 3
	method["write"] = 3
	method["append"] = 3
	method["read"] = 2
	method["set"] = 3
	method["echo"] = 2
	method["find"] = 3
	method["each"] = 3
	method["method"] = 3
	method["exec"] = 3
	regex = regexp.MustCompile(`("[^"]+"|\$?\w+)`)
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

	for n, i := range lines {
		if i[0] == []byte("#")[0] {
			continue
		}
		tokens := regex.FindAll(i, -1)
		for _, v := range tokens {
			fmt.Println(string(v))
		}
		verifyTokenSlice(filePath, tokens, n+1)
		// Insert interpreted instruction to vm

	}
	// If there were no errors:
	// Loop interpreted instruction set
	// Execute each command in sequential order
	// Immediately stop if a command execution results in an error
}

func verifyTokenSlice(filePath string, tokens [][]byte, lineNum int) {
	length := len(tokens)
	action := string(tokens[0])
	reqLength := method[action]
	if reqLength == 0 {
		cmdName := "grep"
		cmdArgs := []string{"--color=always", "-n", "-3", action, filePath}
		cmdOut, _ := exec.Command(cmdName, cmdArgs...).Output()
		log.Println(string(cmdOut))
		log.Fatal("Bad method name: " + action)
	}
	if length < reqLength {
		cmdName := "grep"
		cmdArgs := []string{"--color=always", "-n", "-3", action, filePath}
		cmdOut, _ := exec.Command(cmdName, cmdArgs...).Output()
		log.Println(string(cmdOut))
		log.Println(action, "requires:", reqLength-1, "parameters")
		log.Fatal("Not enough parameters to method: " + action)
	}
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
