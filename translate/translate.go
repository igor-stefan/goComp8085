package translate

import (
	"fmt"
	"strings"

	"github.com/igor-stefan/compiladorAssembly8085/check"
	"github.com/igor-stefan/compiladorAssembly8085/models"
)

const ADDRESS_DATA_SIZE int = 16
const VALUE_DATA_SIZE int = 8

func Opcodesss(opcode string, operand string, regmap map[string]string) (ret string, err error) {
	a := strings.ToLower(operand)
	ret = ""
	err = nil
	if check.IsValidRegister(a) {
		ret = strings.Replace(opcode, "SSS", regmap[a], 1)
	} else {
		err = fmt.Errorf("register %q is invalid. please use registers A through E, H, L or M", operand)
	}
	return
}

func Opcodeddd(opcode string, operand string, regmap map[string]string) (ret string, err error) {
	a := strings.ToLower(operand)
	ret = ""
	err = nil
	if check.IsValidRegister(a) {
		ret = strings.Replace(opcode, "DDD", regmap[a], 1)
	} else {
		err = fmt.Errorf("register %q is invalid. please use registers A through E, H, L or M", operand)
	}
	return
}

func Opcodedddsss(opcode string, operand1 string, operand2 string, regmap map[string]string) (ret string, err error) {
	ret = ""
	err = nil
	var f1 bool = check.IsValidRegister(operand1)
	var f2 bool = check.IsValidRegister(operand2)
	if f1 && f2 {
		opcode = strings.Replace(opcode, "DDD", regmap[strings.ToLower(operand1)], 1)
		ret = strings.Replace(opcode, "SSS", regmap[strings.ToLower(operand2)], 1)
	} else {
		if !f1 && !f2 {
			err = fmt.Errorf("registers %q and %q are invalid. please use registers A through E, H, L or M", operand1, operand2)
		} else if !f2 {
			err = fmt.Errorf("register %q is invalid. please use registers A through E, H, L or M", operand2)
		} else {
			err = fmt.Errorf("register %q is invalid. please use registers A through E, H, L or M", operand1)
		}
	}
	return
}

func Opcoder(opcode string, operand string, regmap map[string]string) (ret string, err error) {
	a := strings.ToLower(operand)
	ret = ""
	err = nil
	if a != "b" && a != "d" {
		err = fmt.Errorf("register %q is invalid. please use registers %q or %q", operand, "B", "D")
	} else {
		ret = strings.Replace(opcode, "R", regmap[a], 1)
	}
	return
}

func Opcoderp(opcode string, operand string, regmap map[string]string) (ret string, err error) {
	a := strings.ToLower(operand)
	ret = ""
	err = nil
	if a != "b" && a != "d" && a != "h" && a != "sp" {
		err = fmt.Errorf("register %q is invalid. please use one of the following registers {%q,%q,%q,%q}", operand, "B", "D", "H", "SP")
	} else {
		ret = strings.Replace(opcode, "RP", regmap[a], 1)
	}
	return
}

func Opcodeccc(opcode string, operand string) (ret string, err error) {
	a := strings.ToLower(operand)
	ret = ""
	err = nil
	if strings.HasPrefix(a, "0b") || strings.HasPrefix(a, "0x") {
		a = a[2:]
	}
	test := []string{"b", "h", "o", "q", "d"}
	for i := 0; i < len(test); i++ {
		if strings.HasSuffix(a, test[i]) {
			a = a[0 : len(a)-1]
			break
		}
	}
	values := []string{"000", "001", "010", "011", "100", "101", "110", "111"}
	var f bool = false
	for i := 0; i < len(values); i++ {
		if a == values[i] {
			f = true
			break
		}
	}
	if !f {
		err = fmt.Errorf("operand %q is invalid. please use values between 000B and 111B", operand)
	} else {
		ret = strings.Replace(opcode, "CCC", a, 1)
	}
	return
}

func Opcodeddddata(opcode string, operand1 string, operand2 string, lbl []models.Label, regmap map[string]string, t map[string]int) (ret [2]string, err error) {
	ret[0] = ""
	ret[1] = ""
	var f1 bool = check.IsValidRegister(operand1)
	err = check.IsValidData(operand2, lbl, VALUE_DATA_SIZE)
	var f2 bool = err == nil
	if f1 && f2 {
		ret[0] = strings.Replace(opcode, "DDD", regmap[strings.ToLower(operand1)], 1)
		ret[1], _ = check.GetBinaryString(operand2, VALUE_DATA_SIZE, lbl, t)
	} else {
		if !f1 && !f2 {
			err = fmt.Errorf("register %q and data %q are both invalid. please use registers A through E, H, L or M. %s", operand1, operand2, err)
		} else if !f2 {
			err = fmt.Errorf("operand %q is invalid. %s", operand2, err)
		} else {
			err = fmt.Errorf("register %q is invalid. please use registers A through E, H, L or M", operand1)
		}
	}
	return
}

func Opcodelhdata(opcode string, operand1 string, lbl []models.Label, t map[string]int) (ret [3]string, err error) {
	ret[0] = ""
	ret[1] = ""
	ret[2] = ""
	err = check.IsValidData(strings.ToLower(operand1), lbl, ADDRESS_DATA_SIZE)
	f1 := err == nil
	if f1 {
		temp, _ := check.GetBinaryString(operand1, ADDRESS_DATA_SIZE, lbl, t)
		ret[0] = opcode
		ret[1] = temp[8:16]
		ret[2] = temp[0:8]
	} else {
		err = fmt.Errorf("operand %q is invalid. %s", operand1, err)
	}
	return
}

func Opcodedata(opcode string, operand1 string, lbl []models.Label, t map[string]int) (ret [2]string, err error) {
	ret[0] = ""
	ret[1] = ""
	err = check.IsValidData(strings.ToLower(operand1), lbl, VALUE_DATA_SIZE)
	f1 := err == nil
	if f1 {
		temp, _ := check.GetBinaryString(operand1, VALUE_DATA_SIZE, lbl, t)
		ret[0] = opcode
		ret[1] = temp
	} else {
		err = fmt.Errorf("operand %q is invalid. %s", operand1, err)
	}
	return
}

func Opcode(mnemonic, operand string) (err error) {
	if operand != "" {
		err = fmt.Errorf("operand for instruction %q must not exist", mnemonic)
	} else {
		err = nil
	}
	return
}
