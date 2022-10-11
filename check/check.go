package check

import (
	"strconv"
	"strings"

	"github.com/igor-stefan/compiladorAssembly8085/models"
)

func IsDirective(a []string, s string) bool {
	for i := 0; i < len(a); i++ {
		if s == a[i] {
			return true
		}
	}
	return false
}

func IsHexData(a string, hasToFit int) bool {
	a = strings.ToLower(a)
	if strings.HasPrefix(a, "0x") {
		if _, err := strconv.ParseUint(a[2:], 16, hasToFit); err == nil {
			return true
		}
	}
	if strings.HasSuffix(a, "h") {
		if _, err := strconv.ParseUint(a[0:len(a)-1], 16, hasToFit); err == nil {
			return true
		}
	}
	return false
}

func IsOctalData(a string, hasToFit int) bool {
	a = strings.ToLower(a)
	if strings.HasSuffix(a, "o") {
		if _, err := strconv.ParseUint(a[0:len(a)-1], 8, hasToFit); err == nil {
			return true
		}
	}
	if strings.HasSuffix(a, "q") {
		if _, err := strconv.ParseUint(a[0:len(a)-1], 8, hasToFit); err == nil {
			return true
		}
	}
	return false

}

func IsDecimalData(a string, hasToFit int) bool {
	a = strings.ToLower(a)
	if strings.HasSuffix(a, "d") {
		if _, err := strconv.ParseUint(a[0:len(a)-1], 10, hasToFit); err == nil {
			return true
		}
	}
	if _, err := strconv.ParseUint(a, 10, hasToFit); err == nil {
		return true
	}
	return false
}

func IsBinaryData(a string, hasToFit int) bool {
	a = strings.ToLower(a)
	if strings.HasPrefix(a, "0b") {
		if _, err := strconv.ParseUint(a[2:], 2, hasToFit); err == nil {
			return true
		}
	}
	if strings.HasSuffix(a, "b") {
		if _, err := strconv.ParseUint(a[0:len(a)-1], 2, hasToFit); err == nil {
			return true
		}
	}
	return false
}

func IsValidRegister(a string) bool {
	a = strings.ToLower(a)
	reg := []string{"a", "b", "c", "d", "e", "f", "h", "l", "m"}
	for i := 0; i < len(reg); i++ {
		if reg[i] == a {
			return true
		}
	}
	return false

}

func IsValidLabel(v []models.Label, a string) bool {
	a = strings.ToLower(a)
	for i := 0; i < len(v); i++ {
		if v[i].Name == a {
			return true
		}
	}
	return false
}

func IsValidAddress(a string, v []models.Label) bool {
	return IsValidLabel(v, a) || IsBinaryData(a, 16) || IsDecimalData(a, 16) || IsHexData(a, 16) || IsOctalData(a, 16)
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
