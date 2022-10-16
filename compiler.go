package main

import (
	"bufio"
	"embed"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"

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
var equTable = make(map[string]int)

// mapping register values
var regsd = make(map[string]string, 10)
var regrp = make(map[string]string, 4)
var regr = make(map[string]string, 2)

var outText []models.Output // all lines of compiled code
var errorText []string      // all errors generated in compilation time
var numErrors int = 0       // counter for errors

//go:embed core/*
var files embed.FS

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
		panic(err)
	}
	defer f1.Close() // remember to close

	f2, err := os.Create("machineCode.txt") // create output file
	if err != nil {
		panic(err)
	}
	defer f2.Close() // remember to close

	infoLogger := log.New(f1, "", 0)
	output := tabwriter.NewWriter(f2, 0, 0, 5, ' ', tabwriter.AlignRight)
	defer output.Flush()

	// outLogger.Printf("Mapa reg r -> %v", regr)

	pattternsFile, err := files.Open("core/patterns.txt") //get all patterns from file
	if err != nil {
		log.Fatal("Error while opening file with patterns, please provide such file")
	}
	defer pattternsFile.Close() // remember to close the file
	patternScanner := bufio.NewScanner(pattternsFile)
	for patternScanner.Scan() {
		lin := patternScanner.Text()
		patterns = append(patterns, strings.Split(lin, " - ")[0])
		// indicator = append(indicator, strings.Split(lin, " - ")[1])
	}

	for _, val := range patterns { //compile the Patterns
		compiledPatterns = append(compiledPatterns, regexp.MustCompile(val))
	}

	cmdSizeFile, err := files.Open("core/cmd_size.txt") //open file with instructions, opcode and size
	if err != nil {
		log.Fatalln("Error while opening file with instruction name, size and opcode, please provide such file")
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
		infoLogger.Printf("Checking line %d...", countLine)
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
				if numPattern > 8 {
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
			infoLogger.Printf("-> Invalid syntax encountered at line %d\n", countLine)
		} else {
			infoLogger.Printf("-> Ok\n")
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
		if mark > 0xffff { // check for address count overflow
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
			valCorrected := val
			if strings.ToLower(ml["mnemonic"]) != "equ" {
				valCorrected = valCorrected[0 : len(valCorrected)-1]
			}
			if l, d := check.IsDuplicateLabel(valCorrected, labels); d {
				numErrors++
				infoLogger.Printf("-> Redefinition of label found")
				errorText = append(errorText, fmt.Sprintf("At line %d: Label %q was already defined in line %d\n", i+1, valCorrected, l))
			} else {
				if strings.ToLower(ml["mnemonic"]) != "equ" {
					labels = append(labels, models.Label{Address: mark, Nline: i, Name: valCorrected})
				} else {
					labels = append(labels, models.Label{Address: mark, Nline: i, Name: valCorrected})
				}
				infoLogger.Printf("-> Valid Label\n")
			}
		}

		if val, exists := ml["mnemonic"]; exists { // checks if mnemonic exists in line
			lowerCaseVal := strings.ToLower(val)
			dir, err := check.IsDirective(directives, lowerCaseVal)
			val1, valid := cmd[lowerCaseVal]
			if valid && err != nil { // check if is an valid mnemonic and not a directive
				mnemonicAdress = append(mnemonicAdress, models.Mnemonic{Start: mark, End: mark + val1.Size - 1, Nline: i, Name: val})
				mark += val1.Size
				infoLogger.Printf("-> Valid Mnemonic\n")
			} else if valid && err == nil { // if it is a valid mnemonic and is a directive
				infoLogger.Printf("-> Valid Directive\n")
				var markChanged bool = false
				switch dir {
				case "org":
					if err = check.IsValidData(ml["op1"], labels, 16); err == nil {
						if hasLabel {
							mnemonicAdress = append(mnemonicAdress, models.Mnemonic{Start: mark, End: mark, Nline: i, Name: val})
						}
						uintVal, err := check.GetIntegerValue(ml["op1"], 16, labels)
						if err == nil {
							markChanged = true
							mark = int(uintVal)
							infoLogger.Printf("-> Memory Address changed by org directive -> New address is 0x%X (%d in base 10)", mark, mark)
						}
					}
				case "db":
					values := strings.Split(ml["op1"], ",")
					var c int = 0
					for j := 0; j < len(values); j++ {
						values[j] = strings.TrimSpace(values[j])
						if values[j] == "" {
							continue
						}
						c++
					}
					if c > 8 {
						err = fmt.Errorf("too much values for %q directive. please use at maximum %d", val, 8)
					} else {
						mnemonicAdress = append(mnemonicAdress, models.Mnemonic{Start: mark, End: mark + c - 1, Nline: i, Name: val})
						markChanged = true
						mark += c
					}
				case "ds":
					if err = check.IsValidData(ml["op1"], labels, 16); err == nil {
						intVal, isEqu := equTable[strings.ToLower(ml["op1"])]
						if !isEqu {
							intVal, err = strconv.Atoi(strings.ToLower(ml["op1"]))
						}
						if intVal == 0 {
							if hasLabel {
								intVal = 1
							} else {
								continue
							}
						}
						mnemonicAdress = append(mnemonicAdress, models.Mnemonic{Start: mark, End: mark + intVal - 1, Nline: i, Name: val})
						markChanged = true
						mark += intVal
					}
				case "equ":
					if err = check.IsValidData(ml["op1"], labels, 16); err == nil {
						intVal, err := check.GetIntegerValue(ml["op1"], 16, labels)
						if err != nil {
							equTable[strings.ToLower(ml["label"])] = equTable[strings.ToLower(ml["op1"])]
							err = nil
						} else {
							equTable[strings.ToLower(ml["label"])] = int(intVal)
						}
						if check.IsValidData(ml["op1"], labels, 8) == nil {
							mnemonicAdress = append(mnemonicAdress, models.Mnemonic{Start: mark, End: mark, Nline: i, Name: val})
						} else {
							mnemonicAdress = append(mnemonicAdress, models.Mnemonic{Start: mark, End: mark + 1, Nline: i, Name: val})
							mark += 2
							markChanged = true
						}
					}
				default:
					panic("directive shouldn't have had a match")
				}
				if err != nil {
					numErrors++
					errorText = append(errorText, fmt.Sprintf("At line %d: %s\n", i+1, err))
					infoLogger.Printf("Error encountered at line %d -> %s\n", i+1, err)
				}
				if !markChanged {
					mark++
				}
			} else {
				infoLogger.Printf("Invalid mnemonic %q at line %d\n", ml["mnemonic"], i+1)
				errorText = append(errorText, fmt.Sprintf("At line %d: invalid mnemonic %q", i+1, ml["mnemonic"]))
				numErrors++
			}
		}
	}

	infoLogger.Printf("\n#Listing labels:\n")
	infoLogger.Printf("\nID  Hex Bin Dec -> Name\n")
	for i, val := range labels {
		infoLogger.Printf("%d. %Xh %08bb %dd -> %s\n", i+1, val.Address, val.Address, val.Address, val.Name)
	}

	infoLogger.Printf("\n#Listing addresses:\n")
	for i, val := range mnemonicAdress {
		infoLogger.Printf("%d. %d to %d -> %Xh to %Xh -> %s\n", i+1, val.Start, val.End, val.Start, val.End, val.Name)
	}

	infoLogger.Printf("\n#Now checking operands and translating to machine code\n")

	for _, val := range mnemonicAdress {
		infoLogger.Printf("Checking line %d...", val.Nline+1)
		lowerCaseValName := strings.ToLower(val.Name)
		if lowerCaseValName == "org" {
			infoLogger.Println("-> Skiped (org directive)")
			continue
		}
		now := cmd[lowerCaseValName] // mnemonic whom operand is being analyzed
		errorsNow := numErrors
		switch now.Translator {
		case 1:
			err := translate.Opcode(linesMatched[val.Nline]["mnemonic"], linesMatched[val.Nline]["op1"])
			if err != nil {
				errorText = append(errorText, fmt.Sprintf("At line %d: %s\n", val.Nline+1, err.Error()))
				numErrors++
			} else {
				outText = append(outText, models.Output{Addr: val.Start, Opcode: now.Opcode})
			}

		case 2:
			code, err := translate.Opcodesss(now.Opcode, linesMatched[val.Nline]["op1"], regsd)
			if err != nil {
				numErrors++
				errorText = append(errorText, fmt.Sprintf("At line %d: %s\n", val.Nline+1, err.Error()))
			} else {
				outText = append(outText, models.Output{Addr: val.Start, Opcode: code})
			}

		case 3:
			code, err := translate.Opcodeddd(now.Opcode, linesMatched[val.Nline]["op1"], regsd)
			if err != nil {
				numErrors++
				errorText = append(errorText, fmt.Sprintf("At line %d: %s\n", val.Nline+1, err.Error()))
			} else {
				outText = append(outText, models.Output{Addr: val.Start, Opcode: code})
			}
		case 4:
			code, err := translate.Opcoderp(now.Opcode, linesMatched[val.Nline]["op1"], regrp)
			if err != nil {
				numErrors++
				errorText = append(errorText, fmt.Sprintf("At line %d: %s\n", val.Nline+1, err.Error()))
			} else {
				outText = append(outText, models.Output{Addr: val.Start, Opcode: code})
			}

		case 5:
			code, err := translate.Opcoder(now.Opcode, linesMatched[val.Nline]["op1"], regr)
			if err != nil {
				numErrors++
				errorText = append(errorText, fmt.Sprintf("At line %d: %s\n", val.Nline+1, err.Error()))
			} else {
				outText = append(outText, models.Output{Addr: val.Start, Opcode: code})
			}

		case 6:
			code, err := translate.Opcodeccc(now.Opcode, linesMatched[val.Nline]["op1"])
			if err != nil {
				numErrors++
				errorText = append(errorText, fmt.Sprintf("At line %d: %s\n", val.Nline+1, err.Error()))
			} else {
				outText = append(outText, models.Output{Addr: val.Start, Opcode: code})
			}

		case 7:
			code, err := translate.Opcodedddsss(now.Opcode, linesMatched[val.Nline]["op1"], linesMatched[val.Nline]["op2"], regsd)
			if err != nil {
				numErrors++
				errorText = append(errorText, fmt.Sprintf("At line %d: %s\n", val.Nline+1, err))
			} else {
				outText = append(outText, models.Output{Addr: val.Start, Opcode: code})
			}

		case 8:
			code, err := translate.Opcodedata(now.Opcode, linesMatched[val.Nline]["op1"], labels, equTable)
			if err != nil {
				numErrors++
				errorText = append(errorText, fmt.Sprintf("At line %d: %s\n", val.Nline+1, err.Error()))
			} else {
				for i := val.Start; i <= val.End; i++ {
					outText = append(outText, models.Output{Addr: i, Opcode: code[i-val.Start]})
				}
			}

		case 9:
			code, err := translate.Opcodeddddata(now.Opcode, linesMatched[val.Nline]["op1"], linesMatched[val.Nline]["op2"], labels, regsd, equTable)
			if err != nil {
				numErrors++
				errorText = append(errorText, fmt.Sprintf("At line %d: %s\n", val.Nline+1, err.Error()))
			} else {
				for i := val.Start; i <= val.End; i++ {
					outText = append(outText, models.Output{Addr: i, Opcode: code[i-val.Start]})
				}
			}

		case 10:
			code, err := translate.Opcodelhdata(now.Opcode, linesMatched[val.Nline]["op1"], labels, equTable)
			if err != nil {
				numErrors++
				errorText = append(errorText, fmt.Sprintf("At line %d: %s\n", val.Nline+1, err))
			} else {
				for i := val.Start; i <= val.End; i++ {
					outText = append(outText, models.Output{Addr: i, Opcode: code[i-val.Start]})
				}
			}

		case 11:
			opcodes, errors := check.CheckDbDirective(linesMatched[val.Nline], labels, equTable)
			if len(errors) > 0 {
				numErrors += len(errors)
				for i := 0; i < len(errors); i++ {
					errorText = append(errorText, fmt.Sprintf("At line %d: %s\n", val.Nline+1, errors[i]))
				}
			} else {
				for i := val.Start; i <= val.End; i++ {
					outText = append(outText, models.Output{Addr: i, Opcode: opcodes[i-val.Start]})
				}
			}

		case 12:
			valueOfEquConstant := equTable[strings.ToLower(linesMatched[val.Nline]["label"])]
			opcode := ""
			if valueOfEquConstant <= 0xff {
				opcode, _ = check.GetFormattedBinaryString(uint64(valueOfEquConstant), 8)
			} else {
				opcode, _ = check.GetFormattedBinaryString(uint64(valueOfEquConstant), 16)
			}
			for i := val.Start; i <= val.End; i++ {
				var k int = 8 * (i - val.Start)
				outText = append(outText, models.Output{Addr: i, Opcode: opcode[0+k : 8+k]})
			}

		case 13:
			for i := val.Start; i <= val.End; i++ {
				outText = append(outText, models.Output{Addr: i, Opcode: "00000000"})
			}

		default:
			outText = append(outText, models.Output{Addr: -1, Opcode: ""})
		}

		if errorsNow != numErrors {
			infoLogger.Printf("-> Error found\n")
		} else {
			infoLogger.Printf("-> Ok\n")
		}
	}

	infoLogger.Printf("\n\n#Now checking for code segment overlap caused by org directive\n")
	addressesUsed := make([]bool, 0xffff)
	var overlap bool = false
	var warnings []string
	var lastIdx int = -1
	for _, mnemonic := range mnemonicAdress {
		infoLogger.Printf("Checking line %d...", mnemonic.Nline+1)
		for j := mnemonic.Start; j <= mnemonic.End; j++ {
			if !addressesUsed[j] {
				addressesUsed[j] = true
			} else {
				overlap = true
				foundLine := false
				idx := 0
				for _, k := range mnemonicAdress {
					for idx = k.Start; idx <= k.End; idx++ {
						if idx == j {
							foundLine = true
							for pos := 0; pos < len(outText); pos++ {
								if outText[pos].Addr == j {
									for pos1 := pos + 1; pos1 < len(outText); pos1++ {
										if outText[pos1].Addr == j {
											outText[pos].Opcode = outText[pos1].Opcode
											outText[pos1] = models.Output{}
										}
									}
								}
							}
							idx = k.Nline + 1
							break
						}
					}
					if foundLine {
						break
					}
				}
				if lastIdx == idx {
					continue
				}
				warnings = append(warnings, fmt.Sprintf("warning at line %d: segment of code starting after %q directive overlaps segment of code starting at line %d\n", mnemonic.Nline, "org", idx))
				infoLogger.Printf("-> Overlap detected\n")
				lastIdx = idx
			}
		}
	}
	if numErrors > 0 {
		c := "s"
		if numErrors < 2 {
			c = ""
		}
		infoLogger.Printf("\nCode compiled with error%s. %d error%s found.", c, numErrors, c)
		log.Printf("Code compiled with error%s. %d error%s found.", c, numErrors, c)
		fmt.Fprintf(output, "Compilation failed. %d error%s found.\n", numErrors, c)
		for _, val := range errorText {
			fmt.Fprintf(output, "%s", val)
		}
	} else {
		infoLogger.Printf("\nCode successfully compiled. No errors found.")
		log.Printf("Code successfully compiled. No errors found.")
		fmt.Fprintf(output, "%s\t%s\t%s\t", "DEC", "HEX", "OPCODE")
		fmt.Fprintf(output, "\n")
		for _, val := range outText {
			if val.Opcode == "" {
				continue
			}
			if val.Addr == -1 {
				fmt.Fprintf(output, "Ttl\tnot\tfound\t")
			} else {
				fmt.Fprintf(output, "%d\t0x%X\t%s\t", val.Addr, val.Addr, val.Opcode)
			}
			fmt.Fprintf(output, "\n")
		}
	}

	if overlap {
		c := ""
		if len(warnings) > 1 {
			c = "s"
		}
		infoLogger.Printf("\nCompilation finished with %d warning%s.", len(warnings), c)
		log.Printf("Compilation finished with %d warning%s.", len(warnings), c)
		fmt.Fprintf(output, "\n")
		for _, w := range warnings {
			fmt.Fprintf(output, "%s", w)
		}
	}
}
