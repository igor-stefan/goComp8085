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
	"github.com/igor-stefan/compiladorAssembly8085/translate"
)

// var indicator []string
var patterns []string
var compiledPatterns []*regexp.Regexp

const MAX_LINES = int(1e4) // constat for max lines in code to be compiled
const CMD_SIZE = int(85)   // constant for map size

var cmd = make(map[string]models.Instruction, CMD_SIZE) // all instructions information
var directives = []string{"db", "org", "ds", "equ"}

// mapping register values
var regsd = make(map[string]string, 10)
var regrp = make(map[string]string, 4)
var regr = make(map[string]string, 2)

var outText []string   // all lines of compiled code
var errorText []string // all errors generated in compilation time
var numErrors int = 0  // counter for errors

func init() {
	// initialize Regr map
	regr["b"] = "0"
	regr["d"] = "1"

	// initialize Regrp map
	regrp["b"] = "00"
	regrp["d"] = "01"
	regrp["h"] = "10"
	regrp["sp"] = "11"

	// initialize Regsd map
	regs := []string{"b", "c", "d", "e", "h", "l", "psw", "sp", "m", "a"}
	values := []string{"000", "001", "010", "011", "100", "101", "110", "110", "110", "111"}
	for i := 0; i < len(regs); i++ {
		regsd[regs[i]] = values[i]
	}
}

func main() {
	// check if it has a file to be compiled
	if len(os.Args) < 2 {
		log.Fatalln("Missing parameter, provide assembly code file name.")
		return
	}

	f1, err := os.Create("compilationLog.txt") // create log file
	if err != nil {
		panic(err) // throw error if case
	}
	defer f1.Close() // remember to close

	f2, err := os.Create("machineCode.txt") // create output file
	if err != nil {
		panic(err) // throw error if case
	}
	defer f2.Close() // remember to close

	infoLogger := log.New(f1, "", 0)
	outLogger := log.New(f2, "", 0)

	// outLogger.Printf("Mapa reg r -> %v", regr)

	pattternsFile, err := os.Open("patterns.txt") //get all patterns from file
	if err != nil {
		log.Fatal("Error while opening file with patterns, please provide such file")
	}
	defer pattternsFile.Close() // remember to close the file after compilation
	patternScanner := bufio.NewScanner(pattternsFile)
	for patternScanner.Scan() {
		lin := patternScanner.Text()
		patterns = append(patterns, strings.Split(lin, " - ")[0])
		// indicator = append(indicator, strings.Split(lin, " - ")[1])
	}

	for _, val := range patterns { //compile the Patterns
		compiledPatterns = append(compiledPatterns, regexp.MustCompile(val))
	}
	// for i, val := range compiledPatterns { //print all patterns
	// 	infoLogger.Printf("%d. %v - %s\n", i, val, indicator[i])
	// }

	cmdSizeFile, err := os.Open("cmd_size.txt") //open file with instructions, opcode and size
	if err != nil {
		log.Fatalln("Error while opening file with instructions, size and opcode, please provide such file")
	}
	defer cmdSizeFile.Close()                       // remember to close the file after compilation
	cmdSizeScanner := bufio.NewScanner(cmdSizeFile) // use constructor to create a scanner
	for cmdSizeScanner.Scan() {
		linSplited := strings.Split(cmdSizeScanner.Text(), ",") // the file is comma separated
		cmdSize, _ := strconv.Atoi(linSplited[1])
		cmdName := linSplited[0]
		cmdOpcode := linSplited[2]
		cmdTranslator, _ := strconv.Atoi(linSplited[3])
		cmd[cmdName] = models.Instruction{Opcode: cmdOpcode, Size: cmdSize, Translator: cmdTranslator}
	}
	// for k, val := range cmd { // logs all info got from file
	// 	infoLogger.Printf("%s -> %v\n", k, val)
	// }

	inputFile, err := os.Open(os.Args[1]) //open file with assembly code

	if err != nil { // handle errors while opening
		log.Fatalf("Error while opening file: %s\n", err)
	}
	defer inputFile.Close()                              // defer to close file as soon as main ends execution
	fileScanner := bufio.NewScanner(inputFile)           //  constructor
	fileScanner.Split(bufio.ScanLines)                   // configure how the scanner behaves
	var countLine int = 0                                // counter of lines for control
	linesMatched := make([]map[string]string, MAX_LINES) // to get all capture groups from regex

	infoLogger.Printf("#Checking pattern match\n\n")
	for fileScanner.Scan() { // read line by line
		countLine++
		// lin := strings.ToLower(fileScanner.Text())
		lin := fileScanner.Text()
		// fmt.Println(strings.TrimRight(lin, "\t ")) // remove white spaces in the right
		m := map[string]string{}
		if lin == "" { // if line is empty, skip
			m["empty"] = "1"
			linesMatched[countLine-1] = m
			continue
		}
		var hasAnyMatch bool = false                    // flag to check if the line has a valid syntax
		for numPattern, val := range compiledPatterns { // check whitch pattern matches with line
			names := val.SubexpNames()      // get capture group names
			matched := val.MatchString(lin) // try to match
			if matched {
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
			numErrors++
			errorText = append(errorText, fmt.Sprintf("At line %d: Invalid syntax\n", countLine))
			infoLogger.Printf("Invalid syntax encountered at line %d\n", countLine)
		}
	}
	// handle first encountered error while reading
	if err := fileScanner.Err(); err != nil {
		log.Fatalf("Error while reading file %s\n", err)
	}
	infoLogger.Printf("Total line count = %d\n\n", countLine)
	infoLogger.Printf("#Listing all lines and mapped values\n\n")
	for i := 0; i < countLine; i++ {
		infoLogger.Printf("%d. %v\n", i+1, linesMatched[i])
	}

	var mnemonicAdress []models.Mnemonic
	var labels []models.Label // armazenate all labels
	var mark int = 0          // marks number of address

	infoLogger.Printf("\n#Now check for mnemonic and label validity\n")
	for i := 0; i < countLine; i++ { // check mnemonic and label validity
		if mark > 0xff { // check for address count overflow
			numErrors++
			infoLogger.Printf("***** CODE OVERFLOWS MEMORY ***** at line %d\n", i)
			errorText = append(errorText, fmt.Sprintf("At line %d: Memory overflow detected!\n", i))
		}
		infoLogger.Print("\n")
		ml := linesMatched[i]
		infoLogger.Printf("Checking line %d...", i+1)
		if _, isEmpty := ml["empty"]; isEmpty {
			infoLogger.Printf("-> Empty Line\n")
			continue
		}
		val, hasLabel := ml["label"] // check for existing label
		if hasLabel {
			labels = append(labels, models.Label{Address: mark, Nline: i, Name: val[0 : len(val)-1]})
			infoLogger.Printf("-> Valid Label\n")
		}
		if val, exists := ml["mnemonic"]; exists { // checks if mnemonic exists in line
			lowerCaseVal := strings.ToLower(val)
			if val1, valid := cmd[lowerCaseVal]; valid { // check if is an valid mnemonic
				mnemonicAdress = append(mnemonicAdress, models.Mnemonic{Start: mark, End: mark + val1.Size - 1, Nline: i, Name: val})
				mark += val1.Size
				infoLogger.Printf("-> Valid Mnemonic\n")
			} else { // if it is not valid mnemonic, it can be a directive
				if check.IsDirective(directives, lowerCaseVal) == nil { // check if it is a valid directive
					// TODO adjust logic for org, db and ds
					if lowerCaseVal != "org" {
						// if (val == "db" || val == "ds") && check.IsDecimalData(ml["op1"], 8) {
						// 	nBytes, err := strconv.Atoi(ml["op1"])
						// 	if err != nil {
						// 		outLogger.Fatalln("db or ds directive number of bytes conversion from str to int failed")
						// 	}
						// 	mark += nBytes
						// 	} else {
						// 		outLogger.Fatalln("db or ds directive number of bytes specified is too large or is invalid")
						// 	}
						mnemonicAdress = append(mnemonicAdress, models.Mnemonic{Start: mark, End: mark, Nline: i, Name: val})
					}
					if lowerCaseVal == "org" && check.IsValidData(ml["op1"], labels, 16) == nil { // check if is org directive to change address counter
						if hasLabel {
							mnemonicAdress = append(mnemonicAdress, models.Mnemonic{Start: mark, End: mark, Nline: i, Name: val})
						}
						uintVal, err := check.GetIntegerValue(ml["op1"], 16, labels)
						if err == nil {
							mark = int(uintVal)
							infoLogger.Printf("-> Memory Address changed by org directive -> New address is 0x%X (%d in base 10)", mark, mark)
						} else {
							numErrors++
							errorText = append(errorText, fmt.Sprintf("At line %d: %s\n", i+1, err))
							infoLogger.Printf("Error encountered at line %d -> %s\n", i+1, err)
						}
					} else {
						mark++
					}
					infoLogger.Printf("-> Valid Directive\n")
					continue
				}
				infoLogger.Printf("Invalid mnemonic %q at line %d\n", ml["mnemonic"], i+1)
				errorText = append(errorText, fmt.Sprintf("At line %d: invalid mnemonic %q", i+1, ml["mnemonic"]))
				numErrors++
			}
		}
	}
	infoLogger.Printf("\n#Listing addresses:\n")
	for i, val := range mnemonicAdress {
		infoLogger.Printf("%d. %d to %d -> %Xh to %Xh -> %s\n", i+1, val.Start, val.End, val.Start, val.End, val.Name)
	}
	infoLogger.Printf("\n#Listing labels:\n")
	for i, val := range labels {
		infoLogger.Printf("%d. %xh -> %s\n", i+1, val.Address, val.Name)
	}

	infoLogger.Printf("\n#Now checking operands and translating into machine code\n")

	for _, val := range mnemonicAdress {
		infoLogger.Printf("Checking line %d...", val.Nline)
		lowerCaseValName := strings.ToLower(val.Name)
		now := cmd[lowerCaseValName] // mnemonic whom operand is being analyzed
		switch now.Translator {
		case 1:
			err := translate.Opcode(linesMatched[val.Nline]["mnemonic"], linesMatched[val.Nline]["op1"])
			if err != nil {
				errorText = append(errorText, fmt.Sprintf("At line %d: %s\n", val.Nline+1, err.Error()))
				numErrors++
			} else {
				outText = append(outText, fmt.Sprintf("%X    %s\n", val.Start, now.Opcode))
			}

		case 2:
			code, err := translate.Opcodesss(now.Opcode, linesMatched[val.Nline]["op1"], regsd)
			if err != nil {
				numErrors++
				errorText = append(errorText, fmt.Sprintf("At line %d: %s\n", val.Nline+1, err.Error()))
			} else {
				outText = append(outText, fmt.Sprintf("%X    %s\n", val.Start, code))
			}

		case 3:
			code, err := translate.Opcodeddd(now.Opcode, linesMatched[val.Nline]["op1"], regsd)
			if err != nil {
				numErrors++
				errorText = append(errorText, fmt.Sprintf("At line %d: %s\n", val.Nline+1, err.Error()))
			} else {
				outText = append(outText, fmt.Sprintf("%X    %s\n", val.Start, code))
			}
		case 4:
			code, err := translate.Opcoderp(now.Opcode, linesMatched[val.Nline]["op1"], regrp)
			if err != nil {
				numErrors++
				errorText = append(errorText, fmt.Sprintf("At line %d: %s\n", val.Nline+1, err.Error()))
			} else {
				outText = append(outText, fmt.Sprintf("%X    %s\n", val.Start, code))
			}

		case 5:
			code, err := translate.Opcoder(now.Opcode, linesMatched[val.Nline]["op1"], regr)
			if err != nil {
				numErrors++
				errorText = append(errorText, fmt.Sprintf("At line %d: %s\n", val.Nline+1, err.Error()))
			} else {
				outText = append(outText, fmt.Sprintf("%X    %s\n", val.Start, code))
			}

		case 6:
			code, err := translate.Opcodeccc(now.Opcode, linesMatched[val.Nline]["op1"])
			if err != nil {
				numErrors++
				errorText = append(errorText, fmt.Sprintf("At line %d: %s\n", val.Nline+1, err.Error()))
			} else {
				outText = append(outText, fmt.Sprintf("%X    %s\n", val.Start, code))
			}

		case 7:
			code, err := translate.Opcodedddsss(now.Opcode, linesMatched[val.Nline]["op1"], linesMatched[val.Nline]["op2"], regsd)
			if err != nil {
				numErrors++
				errorText = append(errorText, fmt.Sprintf("At line %d: %s\n", val.Nline+1, err))
			} else {
				outText = append(outText, fmt.Sprintf("%X    %s\n", val.Start, code))
			}

		case 8:
			code, err := translate.Opcodedata(now.Opcode, linesMatched[val.Nline]["op1"], labels)
			if err != nil {
				numErrors++
				errorText = append(errorText, fmt.Sprintf("At line %d: %s\n", val.Nline+1, err.Error()))
			} else {
				for i := val.Start; i <= val.End; i++ {
					outText = append(outText, fmt.Sprintf("%X    %s\n", i, code[i-val.Start]))
				}
			}

		case 9:
			code, err := translate.Opcodeddddata(now.Opcode, linesMatched[val.Nline]["op1"], linesMatched[val.Nline]["op2"], labels, regsd)
			if err != nil {
				numErrors++
				errorText = append(errorText, fmt.Sprintf("At line %d: %s\n", val.Nline+1, err.Error()))
			} else {
				for i := val.Start; i <= val.End; i++ {
					outText = append(outText, fmt.Sprintf("%X    %s\n", i, code[i-val.Start]))
				}
			}

		case 10:
			code, err := translate.Opcodelhdata(now.Opcode, linesMatched[val.Nline]["op1"], labels)
			// log.Printf("Called by %q at line %d", val.Name, val.Nline)
			// log.Printf("Code = %v\n", code)
			// log.Printf("Err = %v\n", err)
			if err != nil {
				numErrors++
				errorText = append(errorText, fmt.Sprintf("At line %d: %s\n", val.Nline+1, err))
			} else {
				for i := val.Start; i <= val.End; i++ {
					// log.Printf("Address = %d, Value = %s", i, code[i-val.Start])
					outText = append(outText, fmt.Sprintf("%X    %s\n", i, code[i-val.Start]))
				}
			}
		default:
			outText = append(outText, "Translator not found\n")
		}

	}

	if numErrors > 0 {
		c := "s"
		if numErrors < 2 {
			c = ""
		}
		infoLogger.Printf("\nCode compiled with error%s. %d error%s found.", c, numErrors, c)
		log.Printf("\nCode compiled with error%s. %d error%s found.", c, numErrors, c)
		for _, val := range errorText {
			outLogger.Printf("%s", val)
		}
	} else {
		infoLogger.Printf("\nCode successfully compiled. No errors found.")
		log.Printf("\nCode successfully compiled. No errors found.")
		for _, val := range outText {
			outLogger.Printf("%s", val)
		}
	}
	// for _, val := range outText {
	// 	outLogger.Printf("%v", val)
	// }
}
