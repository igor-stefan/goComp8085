package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func main() {
	// open the file
	file, err := os.Open("test.txt")

	//handle errors while opening
	if err != nil {
		log.Fatalf("Error when opening file: %s", err)
	}

	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanWords) //configure how the scanner behaves
	// read line by line
	for fileScanner.Scan() {
		fmt.Println(fileScanner.Text())
	}
	// handle first encountered error while reading
	if err := fileScanner.Err(); err != nil {
		log.Fatalf("Error while reading file: %s", err)
	}

	file.Close()
}
