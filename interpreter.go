package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type vmAction struct {
	method string
	params []string
}

type vm struct {
	actions []vmAction
	state   map[string]string
	groups  map[string][]string
}

func (v *vm) beginExecution() {
	v.state = make(map[string]string)
	v.groups = make(map[string][]string)
	for _, i := range v.actions {
		v.execute(i)
	}
}

func (v *vm) execute(action vmAction) {
	infoPrint("\033[40mExecuting: ", action, "\033[49m")
	switch action.method {
	case "echo":
		for n, i := range action.params {
			if n != 0 {
				fmt.Print(" ")
			}
			if i[0] == '$' {
				fmt.Println(v.state[i])
			} else {
				fmt.Print(strings.Trim(i, "\""))
			}
		}
		fmt.Print("\n")
	case "exec":
		command := action.params[0]
		var variable string
		if len(action.params) > 1 {
			variable = action.params[1]
		}
		command = strings.Trim(command, "\"")
		parts := strings.Split(command, " ")
		cmdName := parts[0]
		var cmdArgs []string
		if len(parts) > 1 {
			for _, i := range parts[1:] {
				if i[0] == '~' {
					cmdArgs = append(cmdArgs, strings.Replace(i, "~", userHome, 1))
				} else {
					cmdArgs = append(cmdArgs, i)
				}
			}
		}
		cmdOut, err := exec.Command(cmdName, cmdArgs...).Output()
		if err != nil {
			fmt.Fprintln(os.Stderr, "There was an error running the command: ", err)
			os.Exit(1)
		}
		if variable == "" {
			fmt.Print(string(cmdOut))
		} else {
			v.state[variable] = string(cmdOut)
		}
	case "set":
		v.state[action.params[0]] = strings.Trim(action.params[1], "\"")
	case "create":
		filename := action.params[0]
		filename = v.processTokenString(filename)
		_, err := os.Stat(filename)
		if err != nil && os.IsNotExist(err) {
			infoPrint("Creating file: " + filename)
			ioutil.WriteFile(filename, []byte(""), 0644)
		} else if err != nil {
			log.Fatal("Could not create file as expected")
		} else {
			log.Fatal("File/folder already exists with that name")
		}
	case "mkdir":
		folder := action.params[0]
		folder = v.processTokenString(folder)
		_, err := os.Stat(folder)
		if err != nil && os.IsNotExist(err) {
			infoPrint("Creating folder: " + folder)
			os.Mkdir(folder, 0700)
		} else if err != nil {
			log.Fatal("Could not create folder as expected")
		} else {
			log.Fatal("File/folder already exists with that name")
		}
	case "cd":
		dir := action.params[0]
		dir = v.processTokenString(dir)
		err := os.Chdir(dir)
		if err != nil {
			log.Fatal("Unable to change directory to: ", dir)
		}
		infoPrint("Changed to directory: " + dir)
	case "rm":
		file := action.params[0]
		file = v.processTokenString(file)
		_, err := os.Stat(file)
		if err != nil && os.IsNotExist(err) {
			log.Fatal("File/directory not found: " + file)
		} else if err != nil {
			log.Fatal("Could not remove file/directory as expected, try again")
		} else {
			infoPrint("Removing file/directory: " + file)
			os.RemoveAll(file)
		}
	case "cp":
		source := action.params[0]
		source = v.processTokenString(source)
		destination := action.params[1]
		destination = v.processTokenString(destination)
		_, err := os.Stat(source)
		if err != nil && os.IsNotExist(err) {
			log.Fatal("Source not found: " + source)
		} else if err != nil {
			log.Fatal("Unexpected error", err)
		} else {
			_, err = os.Stat(destination)
			if err != nil && os.IsNotExist(err) {
				err := exec.Command("cp", []string{"-R", source, destination}...).Run()
				if err != nil {
					log.Fatal(err)
				}
				infoPrint("Copied: " + source + " to: " + destination)
			} else if err != nil {
				log.Fatal("Unexpected error", err)
			} else {
				log.Fatal("Destination already exists, cp can't overwrite")
			}
		}
	case "mv":
		source := action.params[0]
		source = v.processTokenString(source)
		destination := action.params[1]
		destination = v.processTokenString(destination)
		_, err := os.Stat(source)
		if err != nil && os.IsNotExist(err) {
			log.Fatal("Source not found: " + source)
		} else if err != nil {
			log.Fatal("Unexpected error", err)
		} else {
			_, err = os.Stat(destination)
			if err != nil && os.IsNotExist(err) {
				err := exec.Command("mv", []string{source, destination}...).Run()
				if err != nil {
					log.Fatal(err)
				}
				infoPrint("Moved: " + source + " to: " + destination)
			} else if err != nil {
				log.Fatal("Unexpected error", err)
			} else {
				log.Fatal("Destination already exists, mv can't overwrite")
			}
		}
	case "read":
		source := action.params[0]
		source = v.processTokenString(source)
		var variable string
		if len(action.params) > 1 {
			variable = action.params[1]
		}
		_, err := os.Stat(source)
		if err != nil && os.IsNotExist(err) {
			log.Fatal("Source not found: " + source)
		} else if err != nil {
			log.Fatal("Unexpected error", err)
		} else {
			data, err := ioutil.ReadFile(source)
			if err != nil {
				log.Fatal("Unable to read data from file")
			}
			infoPrint("Read data from: " + source)
			if variable == "" {
				fmt.Println(string(data))
			} else {
				v.state[variable] = string(data)
			}
		}
	case "write":
		file := action.params[0]
		file = v.processTokenString(file)
		data := action.params[1]
		data = v.processTokenString(data)
		err := ioutil.WriteFile(file, []byte(data), 0644)
		if err != nil {
			log.Fatal("Unable to write to file:", err)
		}
		infoPrint("Wrote to file: ", file)
	case "append":
		file := action.params[0]
		file = v.processTokenString(file)
		data := action.params[1]
		data = v.processTokenString(data)
		f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			log.Fatal(err)
		}

		defer f.Close()

		if _, err = f.WriteString(data); err != nil {
			log.Fatal(err)
		}
		infoPrint("Appended to file: ", file)
	case "find":
		findText := action.params[0]
		findText = v.processTokenString(findText)
		groupVar := action.params[1]
		if groupVar[0] == '[' && groupVar[1] == ']' {
			groupVar = groupVar[2:]
			curDir, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			cmdOut, err := exec.Command("find", curDir, "-iname", findText).Output()
			if err != nil {
				fmt.Fprintln(os.Stderr, "There was an error running the command: ", err)
				os.Exit(1)
			}
			files := bytes.Split(cmdOut, []byte("\n"))
			var found []string
			for _, i := range files {
				if len(i) > 0 {
					found = append(found, string(i))
				}
			}
			v.groups[groupVar] = found
		} else {
			log.Fatal("You need to use a group variable for find commands")
		}
	case "each":
		groupVar := action.params[0]
		if groupVar[0] == '[' && groupVar[1] == ']' {
			groupVar = groupVar[2:]
			command := action.params[1]
			command = strings.Trim(command, "\"")
			parts := strings.Split(command, " ")
			cmdName := parts[0]
			var cmdArgs []string
			if len(parts) > 1 {
				for _, i := range parts[1:] {
					if i[0] == '~' {
						cmdArgs = append(cmdArgs, strings.Replace(i, "~", userHome, 1))
					} else {
						cmdArgs = append(cmdArgs, i)
					}
				}
			}
			index := -1
			for n, v := range cmdArgs {
				if v == "$$" {
					index = n
				}
			}

			files := v.groups[groupVar]
			for _, i := range files {
				infoPrint(i)
				if index == -1 {
					cmdOut, err := exec.Command(cmdName, cmdArgs...).Output()
					if err != nil {
						fmt.Fprintln(os.Stderr, "There was an error running the command: ", err)
						os.Exit(1)
					}
					fmt.Println(string(cmdOut))
				} else {
					copyArgs := make([]string, len(cmdArgs))
					copy(copyArgs, cmdArgs)
					copyArgs[index] = i
					cmdOut, err := exec.Command(cmdName, copyArgs...).Output()
					if err != nil {
						fmt.Fprintln(os.Stderr, "There was an error running the command: ", err)
						os.Exit(1)
					}
					fmt.Println(string(cmdOut))
				}
			}
		} else {
			log.Fatal("You need to use a group variable for find commands")
		}
	default:
		fmt.Println(action)
	}
	fmt.Println("")
}

func (v *vm) processTokenString(token string) string {
	if token[0] == '$' {
		token = v.state[token]
	}
	if token[0] == '"' {
		var err error
		token, err = strconv.Unquote(token)
		if err != nil {
			log.Fatal(err)
		}
	}
	if token[0] == '~' {
		token = strings.Replace(token, "~", userHome, 1)
	}
	return token
}

var (
	version  = "0.0.1"
	info     *log.Logger
	method   = make(map[string]int)
	regex    *regexp.Regexp
	userHome string
)

func init() {
	log.SetFlags(0)
	method["create"] = 2
	method["mkdir"] = 2
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
	method["exec"] = 2
	regex = regexp.MustCompile(`("[^"]+"|[\$\[]?\]?[\w\/~\.]+)`)
	userHome = os.Getenv("HOME")
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
	infoPrint(versionInfo())

	checkFileExists(filePath)

	contents := getContents(filePath)

	lines := bytes.Split(contents, []byte("\n"))

	var run vm

	for n, i := range lines {
		if len(i) > 0 {
			if i[0] == []byte("#")[0] {
				continue
			}
			tokens := regex.FindAll(i, -1)
			action := verifyTokenSlice(filePath, tokens, n+1)
			run.actions = append(run.actions, action)
		}
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
	infoPrint("Loaded contents of ", filePath)
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

func infoPrint(message ...interface{}) {
	info.Print("\033[32m")
	info.Print(message...)
	info.Println("\033[39m")
}

func versionInfo() string {
	return ("GOSH INTERPRETER VER " + version)
}
