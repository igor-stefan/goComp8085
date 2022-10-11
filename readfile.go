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

var cmdSize = make(map[string]int, CMD_SIZE)
var directives = []string{"db", "org", "ds", "equ"}

func main() {
	// check if it has a file to be compiled
	if len(os.Args) < 2 {
		fmt.Println("Missing parameter, provide file name.")
		return
	}

	pattternsFile, err := os.Open("patterns.txt")
	if err != nil {
		log.Fatalln("Error opening file with patterns, please provide such file")
	}
	defer pattternsFile.Close()
	patternScanner := bufio.NewScanner(pattternsFile)

	for patternScanner.Scan() {
		lin := patternScanner.Text()
		patterns = append(patterns, strings.Split(lin, " - ")[0])
		indicator = append(indicator, strings.Split(lin, " - ")[1])
	}
	for _, val := range patterns {
		compiledPatterns = append(compiledPatterns, regexp.MustCompile(val))
	}
	// for i, val := range compiledPatterns {
	// 	fmt.Printf("%d. %v - %s\n", i, val, indicator[i])
	// }
	fmt.Println()

	cmdSizeFile, err := os.Open("cmd_size.txt")
	if err != nil {
		log.Fatalln("Error opening file with command size, please provide such file")
	}
	defer cmdSizeFile.Close()
	cmdSizeScanner := bufio.NewScanner(cmdSizeFile)
	for cmdSizeScanner.Scan() {
		linSplited := strings.Split(cmdSizeScanner.Text(), ",")
		intVal, _ := strconv.Atoi(linSplited[1])
		cmdSize[linSplited[0]] = intVal
	}
	// for key, val := range cmdSize {
	// 	fmt.Printf("%s -> %d\n", key, val)
	// }
	//open the file
	inputFile, err := os.Open(os.Args[1])

	// handle errors while opening
	if err != nil {
		log.Fatalf("Error when opening file: %s\n", err)
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
				// fmt.Printf("\nnames -> %v\n", names)
				// fmt.Printf("result -> %v\n", result)
				// fmt.Printf("mapa M -> %v\n", m)
			}
		}
		if !hasAnyMatch {
			log.Fatalf("Invalid syntax at line %d\n", countLine)
		}
	}
	// handle first encountered error while reading
	if err := fileScanner.Err(); err != nil {
		log.Fatalf("Error while reading file %s\n", err)
	}
	fmt.Printf("\nTotal de linhas = %d\n", countLine)
	for i := 0; i < countLine; i++ {
		fmt.Printf("%d. %v\n", i, linesMatched[i])
	}

	var mnemonicAdress []models.Mnemonic
	var labels []models.Label
	var mark int = 0

	fmt.Printf("\n checando as linhas\n")
	for i := 0; i < countLine; i++ {
		ml := linesMatched[i]
		fmt.Printf("%d. %v\n", i, ml)
		if _, isEmpty := ml["empty"]; isEmpty {
			fmt.Printf("linha %d vazia\n", i)
			continue
		}
		if val, ok := ml["label"]; ok {
			labels = append(labels, models.Label{Address: mark, Nline: i, Name: val[:len(val)-1]})
		}
		if val, ok := ml["mnemonic"]; ok { // checks if mnemonic exists
			if val1, ok1 := cmdSize[val]; ok1 { // check if is an valid mnemonic
				mnemonicAdress = append(mnemonicAdress, models.Mnemonic{Start: mark, End: mark + val1 - 1, Nline: i, Name: val})
				mark += val1
				// if dir && val == "org" && validOp1(ml["op1"]) {
				// 	mark = val1
				// }
				fmt.Printf("valido mnemonico\n")
			} else {
				if dir := check.IsDirective(directives, val); dir {
					mnemonicAdress = append(mnemonicAdress, models.Mnemonic{Start: mark, End: mark, Nline: i, Name: val})
					mark++
					continue
				}
				log.Fatalf("Invalid mnemonic at line %d\n", i+1)
			}
		}
	}
	fmt.Printf("\nNúmero de endereços:\n")
	for i, val := range mnemonicAdress {
		fmt.Printf("%d. %d até %d -> %s\n", i, val.Start, val.End, val.Name)
	}
	fmt.Printf("\nRelacao de labels:\n")
	for i, val := range labels {
		fmt.Printf("%d. %xh -> %s\n", i, val.Address, val.Name)
	}

	fmt.Printf("\nTeste de funcao\n")
	fmt.Printf("%v\n", check.IsDecimalData("258"))
}
