package check

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/igor-stefan/compiladorAssembly8085/models"
)

const reghex string = `(?i)(?:\b[a-f0-9]+h{1}\b)|(?:\b(?:0x){1}[a-f0-9]+\b)`
const regoct string = `(?i)(?:\b[0-7]+(?:o|q){1}\b)`
const regdec string = `(?i)(?:\b[0-9]+d{1}\b)|\b[0-9]+\b`
const regbin string = `(?i)(?:\b[0-1]+b{1}\b)|(?:\b(?:0b){1}[0-1]+\b)`
const reglbl string = `^\w+$`

var regh, rego, regd, regb, regl *regexp.Regexp

func init() {
	regh = regexp.MustCompile(reghex)
	rego = regexp.MustCompile(regoct)
	regd = regexp.MustCompile(regdec)
	regb = regexp.MustCompile(regbin)
	regl = regexp.MustCompile(reglbl)
}

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

func IsValidHexData(s string, hasToFit int) (err error) {
	a := strings.ToLower(s)
	a = CutStringForParse(a)
	err = nil
	_, err = strconv.ParseUint(a, 16, hasToFit)
	if err != nil {
		err = GetError(err.(*strconv.NumError), err, s, hasToFit)
	}
	return
}

func IsValidOctalData(s string, hasToFit int) (err error) {
	a := strings.ToLower(s)
	a = CutStringForParse(a)
	err = nil
	_, err = strconv.ParseUint(a, 8, hasToFit)
	if err != nil {
		err = GetError(err.(*strconv.NumError), err, s, hasToFit)
	}
	return
}

func IsValidDecimalData(s string, hasToFit int) (err error) {
	a := strings.ToLower(s)
	a = CutStringForParse(a)
	err = nil
	_, err = strconv.ParseUint(a, 10, hasToFit)
	if err != nil {
		err = GetError(err.(*strconv.NumError), err, s, hasToFit)
	}
	return
}

func IsValidBinaryData(s string, hasToFit int) (err error) {
	a := strings.ToLower(s)
	a = CutStringForParse(a)
	err = nil
	_, err = strconv.ParseUint(a, 2, hasToFit)
	if err != nil {
		err = GetError(err.(*strconv.NumError), err, s, hasToFit)
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

func IsValidLabel(v []models.Label, s string) (idx int) {
	a := strings.ToLower(s)
	idx = -1
	for i := 0; i < len(v); i++ {
		if strings.ToLower(v[i].Name) == a {
			idx = i
			break
		}
	}
	return
}

func IsValidData(s string, v []models.Label, bitSize int) (err error) {
	a := strings.ToLower(s)
	err = nil
	if regh.MatchString(a) {
		err = IsValidHexData(s, bitSize)
	} else if regb.MatchString(a) {
		err = IsValidBinaryData(s, bitSize)
	} else if rego.MatchString(a) {
		err = IsValidOctalData(s, bitSize)
	} else if regd.MatchString(a) {
		err = IsValidDecimalData(s, bitSize)
	} else if regl.MatchString(a) && IsValidLabel(v, a) > -1 {
		err = nil
	} else {
		err = fmt.Errorf("operand %q is invalid", s)
	}
	return
}

func GetIntegerValue(s string, hasToFit int, lbl []models.Label) (x uint64, err error) {
	a := strings.ToLower(s)
	x = 0
	base, err := GetBase(a, lbl)
	if err != nil {
		return
	}
	a = CutStringForParse(a)
	x, err = strconv.ParseUint(a, base, hasToFit)
	if err == nil {
		return
	} else {
		err = GetError(err.(*strconv.NumError), err, s, hasToFit)
	}
	return
}

func GetBinaryString(a string, bitSize int, lbl []models.Label) (ret string, err error) {
	s := strings.ToLower(a)
	err = nil
	ret = ""
	base, err := GetBase(s, lbl)
	if err != nil {
		return
	}
	if base == 20 {
		idx := IsValidLabel(lbl, a)
		ret, err = GetFormattedBinaryString(uint64(lbl[idx].Address), bitSize)
		return
	}
	s = CutStringForParse(s)
	x, err := strconv.ParseUint(s, base, bitSize)
	if err != nil {
		err = GetError(err.(*strconv.NumError), err, a, bitSize)
		return
	} else {
		ret, err = GetFormattedBinaryString(x, bitSize)
	}
	return
}

func GetBase(s string, lbl []models.Label) (base int, err error) {
	err = nil
	base = -1
	s = strings.ToLower(s)
	if regb.MatchString(s) {
		base = 2
	} else if rego.MatchString(s) {
		base = 8
	} else if regd.MatchString(s) {
		base = 10
	} else if regh.MatchString(s) {
		base = 16
	} else if regl.MatchString(s) && IsValidLabel(lbl, s) > -1 {
		base = 20
	} else {
		err = fmt.Errorf("the following operand is invalid -> %q", s)
	}
	return
}

func GetError(numError *strconv.NumError, err error, s string, bitWidth int) (err1 error) {
	err1 = nil
	if numError, ok := err.(*strconv.NumError); ok {
		if numError.Err == strconv.ErrRange {
			err1 = fmt.Errorf("value %q overflows %d bit width", s, bitWidth)
		} else if numError.Err == strconv.ErrSyntax {
			err1 = fmt.Errorf("value %q has invalid syntax", s)
		}
	}
	return
}

func CutStringForParse(s string) string {
	s = strings.ToLower(s)
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
	return s
}

func GetFormattedBinaryString(x uint64, bitWidth int) (ret string, err error) {
	err = nil
	ret = ""
	if bitWidth == 8 {
		ret = fmt.Sprintf("%08b", x)
	} else if bitWidth == 16 {
		ret = fmt.Sprintf("%016b", x)
	} else {
		err = fmt.Errorf("size of bit width (%d) is invalid. please choose 16 or 8", bitWidth)
	}
	return
}
