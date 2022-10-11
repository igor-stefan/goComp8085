package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/igor-stefan/compiladorAssembly8085/check"
	"github.com/igor-stefan/compiladorAssembly8085/models"
)

var patterns []string
var indicator []string
var compiledPatterns []*regexp.Regexp

const MAX_LINES = int(1e4)
const CMD_SIZE = int(85)

var cmd = make(map[string]models.Instruction, CMD_SIZE)
var directives = []string{"db", "org", "ds", "equ"}

func main() {
	// check if it has a file to be compiled
	if len(os.Args) < 2 {
		fmt.Println("Missing parameter, provide file name.")
		return
	}

	f1, err := os.Create("file1.txt")
	if err != nil {
		panic(err)
	}
	defer f1.Close()

	f2, err := os.Create("file2.txt")
	if err != nil {
		panic(err)
	}
	defer f2.Close()

	infoLogger := log.New(f1, "", 0)
	outLogger := log.New(f2, "", 0)

	pattternsFile, err := os.Open("patterns.txt") //get all patterns
	if err != nil {
		outLogger.Fatalln("Error opening file with patterns, please provide such file")
	}
	defer pattternsFile.Close()
	patternScanner := bufio.NewScanner(pattternsFile)

	for patternScanner.Scan() {
		lin := patternScanner.Text()
		patterns = append(patterns, strings.Split(lin, " - ")[0])
		indicator = append(indicator, strings.Split(lin, " - ")[1])
	}
	for _, val := range patterns { //compile the Patterns
		compiledPatterns = append(compiledPatterns, regexp.MustCompile(val))
	}
	for i, val := range compiledPatterns { //print all patterns
		outLogger.Printf("%d. %v - %s\n", i, val, indicator[i])
	}
	cmdSizeFile, err := os.Open("cmd_size.txt") //open file with instructions
	if err != nil {
		outLogger.Fatalln("Error opening file with command size, please provide such file")
	}
	defer cmdSizeFile.Close()
	cmdSizeScanner := bufio.NewScanner(cmdSizeFile)
	for cmdSizeScanner.Scan() {
		linSplited := strings.Split(cmdSizeScanner.Text(), ",")
		cmdSize, _ := strconv.Atoi(linSplited[1])
		cmdName := linSplited[0]
		cmdOpcode := linSplited[2]
		cmd[cmdName] = models.Instruction{Opcode: cmdOpcode, Size: cmdSize}
	}
	for k, val := range cmd {
		fmt.Printf("%s -> %v\n", k, val)
	}
	//open the file
	inputFile, err := os.Open(os.Args[1])

	// handle errors while opening
	if err != nil {
		outLogger.Fatalf("Error when opening file: %s\n", err)
	}
	defer inputFile.Close() // defer to close file as soon as main ends execution

	fileScanner := bufio.NewScanner(inputFile)           //  constructor
	fileScanner.Split(bufio.ScanLines)                   // configure how the scanner behaves
	var countLine int = 0                                // counter of lines to display error messages if any
	linesMatched := make([]map[string]string, MAX_LINES) // to construct
	for fileScanner.Scan() {                             // read line by line
		countLine++
		lin := strings.ToLower(fileScanner.Text()) // lowercase all string
		fmt.Println(strings.TrimRight(lin, "\t ")) // remove white spaces in the right
		if lin == "" {                             // if line is empty, skip
			// fmt.Print("skip empty line\n")
			m := map[string]string{}
			m["empty"] = "1"
			linesMatched[countLine-1] = m
			continue
		}
		var hasAnyMatch bool = false // flag to check if the line has a valid syntax
		for numPattern, val := range compiledPatterns {
			names := val.SubexpNames()
			matched := val.MatchString(lin)
			if matched {
				m := map[string]string{}
				hasAnyMatch = true
				if numPattern > 6 {
					m["empty"] = "1"
					linesMatched[countLine-1] = m
					continue
				}
				result := val.FindAllStringSubmatch(lin, -1)
				for j := 1; j < len(names); j++ {
					m[names[j]] = result[0][j]
				}
				linesMatched[countLine-1] = m
			}
		}
		if !hasAnyMatch {
			outLogger.Fatalf("Invalid syntax at line %d\n", countLine)
		}
	}
	// handle first encountered error while reading
	if err := fileScanner.Err(); err != nil {
		outLogger.Fatalf("Error while reading file %s\n", err)
	}
	infoLogger.Printf("Total line count = %d\n", countLine)
	for i := 0; i < countLine; i++ {
		infoLogger.Printf("%d. %v\n", i+1, linesMatched[i])
	}

	var mnemonicAdress []models.Mnemonic
	var labels []models.Label
	var mark int = 0

	infoLogger.Printf("\nNow check for mnemonic and label validity\n")
	for i := 0; i < countLine; i++ { // check mnemonic and label validity
		infoLogger.Print("\n")
		ml := linesMatched[i]
		infoLogger.Printf("Checking line %d...", i+1)
		if _, isEmpty := ml["empty"]; isEmpty {
			infoLogger.Printf("-> Empty Line\n")
			continue
		}
		if val, ok := ml["label"]; ok {
			labels = append(labels, models.Label{Address: mark, Nline: i, Name: val[:len(val)-1]})
			infoLogger.Printf("-> Valid Label\n")
		}
		if val, ok := ml["mnemonic"]; ok { // checks if mnemonic exists in line
			if val1, ok1 := cmd[val]; ok1 { // check if is an valid mnemonic
				mnemonicAdress = append(mnemonicAdress, models.Mnemonic{Start: mark, End: mark + val1.Size - 1, Nline: i, Name: val})
				mark += val1.Size
				infoLogger.Printf("-> Valid Mnemonic\n")
			} else {
				if dir := check.IsDirective(directives, val); dir {
					mnemonicAdress = append(mnemonicAdress, models.Mnemonic{Start: mark, End: mark, Nline: i, Name: val})
					mark++
					if dir && val == "org" && check.IsValidAddress(ml["op1"], labels) {
						mark = check.GetIntegerValue(ml["op1"], 16)
					}
					infoLogger.Printf("-> Valid Directive\n")
					continue
				}
				outLogger.Fatalf("Invalid mnemonic at line %d\n", i+1)
			}
		}
	}
	infoLogger.Printf("\nListing adresses:\n")
	for i, val := range mnemonicAdress {
		infoLogger.Printf("%d. %d até %d -> %s\n", i+1, val.Start, val.End, val.Name)
	}
	infoLogger.Printf("\nListing labels:\n")
	for i, val := range labels {
		infoLogger.Printf("%d. %xh -> %s\n", i+1, val.Address, val.Name)
	}

	infoLogger.Printf("\nTeste de funcao\n")
	infoLogger.Printf("%v\n", check.IsHexData("5", 2))
}
