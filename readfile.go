package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

var patterns []string
var indicator []string
var compiledPatterns []*regexp.Regexp

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
	for i, val := range compiledPatterns {
		fmt.Printf("%d. %v - %s\n", i, val, indicator[i])
	}
	fmt.Println()
	//open the file
	inputFile, err := os.Open(os.Args[1])

	// handle errors while opening
	if err != nil {
		log.Fatalf("Error when opening file: %s\n", err)
	}
	defer inputFile.Close() // defer to close file as soon as main ends execution

	fileScanner := bufio.NewScanner(inputFile) //  constructor
	fileScanner.Split(bufio.ScanLines)         // configure how the scanner behaves
	// read line by line
	for fileScanner.Scan() {
		lin := fileScanner.Text()
		fmt.Print(strings.TrimRight(lin, "\t "))
		if lin == "" {
			fmt.Print("skip empty line\n")
			continue
		}
		var f bool = false
		for i, val := range compiledPatterns {
			matched := val.MatchString(lin)
			if matched {
				f = true
				fmt.Print(" ---> match com padrao ", i, " ---> ", indicator[i], "\n")
			}
		}
		if !f {
			fmt.Println(" ---> nao reconhecido")
		}
	}
	// handle first encountered error while reading
	if err := fileScanner.Err(); err != nil {
		log.Fatalf("Error while reading file %s\n", err)
	}
}
