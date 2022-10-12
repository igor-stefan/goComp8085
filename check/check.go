package check

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/igor-stefan/compiladorAssembly8085/models"
)

func IsDirective(a []string, s string) (err error) {
	err = nil
	f := false
	for i := 0; i < len(a); i++ {
		if s == a[i] {
			f = true
			break
		}
	}
	if !f {
		err = fmt.Errorf("argument %q doesn't match with any directive", s)
	}
	return
}

func IsHexData(s string, hasToFit int) (err error) {
	a := strings.ToLower(s)
	f1, f2 := false, false
	if strings.HasPrefix(a, "0x") {
		if _, err := strconv.ParseUint(a[2:], 16, hasToFit); err == nil {
			f1 = true
		}
	}
	if strings.HasSuffix(a, "h") {
		if _, err := strconv.ParseUint(a[0:len(a)-1], 16, hasToFit); err == nil {
			f2 = true
		}
	}
	if f1 || f2 {
		err = nil
	} else {
		err = fmt.Errorf("%q is not valid hexadecimal data", s)
	}
	return
}

func IsOctalData(s string, hasToFit int) (err error) {
	a := strings.ToLower(s)
	err = nil
	f1, f2 := false, false
	if strings.HasSuffix(a, "o") {
		if _, err := strconv.ParseUint(a[0:len(a)-1], 8, hasToFit); err == nil {
			f1 = true
		}
	}
	if strings.HasSuffix(a, "q") {
		if _, err := strconv.ParseUint(a[0:len(a)-1], 8, hasToFit); err == nil {
			f2 = true
		}
	}
	if !f1 && !f2 {
		err = fmt.Errorf("%q is not valid octal data", s)
	}
	return
}

func IsDecimalData(s string, hasToFit int) (err error) {
	a := strings.ToLower(s)
	err = nil
	f1, f2 := false, false
	if strings.HasSuffix(a, "d") {
		if _, err := strconv.ParseUint(a[0:len(a)-1], 10, hasToFit); err == nil {
			f1 = true
		}
	}
	if _, err := strconv.ParseUint(a, 10, hasToFit); err == nil {
		f2 = true
	}
	if !f1 && !f2 {
		err = fmt.Errorf("%q is not valid decimal data", s)
	}
	return
}

func IsBinaryData(s string, hasToFit int) (err error) {
	a := strings.ToLower(s)
	err = nil
	f1, f2 := false, false
	if strings.HasPrefix(a, "0b") {
		if _, err := strconv.ParseUint(a[2:], 2, hasToFit); err == nil {
			f1 = true
		}
	}
	if strings.HasSuffix(a, "b") {
		if _, err := strconv.ParseUint(a[0:len(a)-1], 2, hasToFit); err == nil {
			f2 = true
		}
	}
	if !f1 && !f2 {
		err = fmt.Errorf("%q is not valid binary data", s)
	}
	return
}

func IsValidRegister(a string) bool {
	a = strings.ToLower(a)
	reg := []string{"a", "b", "c", "d", "e", "h", "l", "m"}
	for i := 0; i < len(reg); i++ {
		if reg[i] == a {
			return true
		}
	}
	return false
}

func IsValidLabel(v []models.Label, s string) (err error) {
	a := strings.ToLower(s)
	f1 := false
	err = nil
	for i := 0; i < len(v); i++ {
		if v[i].Name == a {
			f1 = true
			break
		}
	}
	if !f1 {
		err = fmt.Errorf("couldn't identify %q as label", s)
	}
	return
}

func IsValidData(a string, v []models.Label, bitSize int) (err error) {
	f1, f2, f3, f4, f5 := false, false, false, false, false
	f1 = IsValidLabel(v, a) == nil
	f2 = IsBinaryData(a, bitSize) == nil
	f3 = IsDecimalData(a, bitSize) == nil
	f4 = IsHexData(a, bitSize) == nil
	f5 = IsOctalData(a, bitSize) == nil
	if f1 || f2 || f3 || f4 || f5 {
		err = nil
	} else {
		err = fmt.Errorf("operand %q is an invalid", a)
	}
	return
}

func GetIntegerValue(a string, hasToFit int) (x int) {
	a = strings.ToLower(a)
	x = -1
	if strings.HasPrefix(a, "0x") || strings.HasPrefix(a, "0b") {
		if x, err := strconv.ParseInt(a[2:], 10, hasToFit); err == nil {
			return int(x)
		}
	}
	if strings.HasSuffix(a, "h") || strings.HasSuffix(a, "o") || strings.HasSuffix(a, "q") || strings.HasSuffix(a, "v") {
		if x, err := strconv.ParseInt(a[0:len(a)-1], 10, hasToFit); err == nil {
			return int(x)
		}
	}
	if x, err := strconv.ParseInt(a, 10, hasToFit); err == nil {
		return int(x)
	}
	return x
}

func GetBinaryString(s string) (ret string, err error) {
	s = strings.ToLower(s)
	err = nil
	ret = ""
	base, err := GetBase(s, 8)
	if err != nil {
		return
	}
	if strings.HasPrefix(s, "0b") || strings.HasPrefix(s, "0x") {
		s = s[2:]
	}
	test := []string{"b", "h", "o", "q", "d"}
	for i := 0; i < len(test); i++ {
		if strings.HasSuffix(s, test[i]) {
			s = s[0 : len(s)-1]
			break
		}
	}
	x, err := strconv.ParseUint(s, base, 8)
	if err == nil {
		ret = fmt.Sprintf("%08b", x)
	}
	return
}

func GetBase(s string, hasToFit int) (base int, err error) {
	err = nil
	base = -1
	s = strings.ToLower(s)
	if IsBinaryData(s, hasToFit) == nil {
		base = 2
	} else if IsDecimalData(s, hasToFit) == nil {
		base = 10
	} else if IsOctalData(s, hasToFit) == nil {
		base = 8
	} else if IsHexData(s, hasToFit) == nil {
		base = 16
	}
	if base == -1 {
		err = fmt.Errorf("the following operand is invalid -> %q", s)
	}
	return
}
