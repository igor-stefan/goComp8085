package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func main() {
	// check if it has a file to be compiled
	if len(os.Args) < 2 {
		fmt.Println("Missing parameter, provide file name.")
		return
	}

	// open the file
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
		fmt.Println(fileScanner.Text())
	}
	// handle first encountered error while reading
	if err := fileScanner.Err(); err != nil {
		log.Fatalf("Error while reading file %s\n", err)
	}
}
