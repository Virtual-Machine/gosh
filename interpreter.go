package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type vmAction struct {
	method string
	params []string
}

type vm struct {
	actions []vmAction
	state   map[string]string
}

func (v *vm) beginExecution() {
	for _, i := range v.actions {
		v.execute(i)
	}
}

func (v *vm) execute(action vmAction) {
	info.Println("\033[40mExecuting:", action, "\033[49m")
	switch action.method {
	case "echo":
		for n, i := range action.params {
			if n != 0 {
				fmt.Print(" ")
			}
			fmt.Print(strings.Trim(i, "\""))
		}
		fmt.Print("\n")
	case "exec":
		command := action.params[0]
		// var variable string
		if len(action.params) > 1 {
			// variable = action.params[1]
		}
		command = strings.Trim(command, "\"")
		parts := strings.Split(command, " ")
		cmdName := parts[0]
		var cmdArgs []string
		if len(parts) > 1 {
			cmdArgs = parts[1:]
		}
		cmdOut, err := exec.Command(cmdName, cmdArgs...).Output()
		if err != nil {
			fmt.Fprintln(os.Stderr, "There was an error running the command: ", err)
			os.Exit(1)
		}
		fmt.Print(string(cmdOut))
	default:
		fmt.Println(action)
	}
	fmt.Println("")
}

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
	method["exec"] = 2
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

	var run vm

	for n, i := range lines {
		if i[0] == []byte("#")[0] {
			continue
		}
		tokens := regex.FindAll(i, -1)
		action := verifyTokenSlice(filePath, tokens, n+1)
		run.actions = append(run.actions, action)
	}

	run.beginExecution()
}

func verifyTokenSlice(filePath string, tokens [][]byte, lineNum int) vmAction {
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
	var parameters []string
	for _, i := range tokens[1:] {
		parameters = append(parameters, string(i))
	}
	return vmAction{method: action, params: parameters}
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
