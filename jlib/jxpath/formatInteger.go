package jxpath

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"

	rom "github.com/brandenc40/romannumeral"
	"github.com/divan/num2words"
	"github.com/samber/lo"
)

var (
	dfForInteger = NewDecimalFormat()
	dividerUnit  = []string{"thousand", "million", "billion", "trillion", "quadrillion", "quintillion", "sextillion"}
	ordinalMap   = map[int]string{
		0:  "zeroth",
		1:  "first",
		2:  "second",
		3:  "third",
		4:  "fourth",
		5:  "fifth",
		6:  "sixth",
		7:  "seventh",
		8:  "eighth",
		9:  "ninth",
		10: "tenth",
		11: "eleventh",
		12: "twelfth",
		13: "thirteenth",
		14: "fourteenth",
		15: "fifteenth",
		16: "sixteenth",
		17: "seventeenth",
		18: "eighteenth",
		19: "nineteenth",
	}
)

// 如果传入 twenty-one， 返回 twenty-first
// 如果传入 one， 返回 first
func getStringOrdinal(str string) string {
	if cardinal2ordinal[str] != "" {
		return cardinal2ordinal[str]
	}
	var strSegs []string
	if strings.Contains(str, "-") {
		strSegs = strings.Split(str, "-")
		if len(strSegs) == 2 {
			strSegs[1] = cardinal2ordinal[strSegs[1]]
			return strings.Join(strSegs, "-")
		}
	}
	if strings.Contains(str, " ") {
		strSegs = strings.Split(str, " ")
		if len(strSegs) == 2 {
			strSegs[1] = cardinal2ordinal[strSegs[1]]
			return strings.Join(strSegs, " ")
		}
	}

	return "unknown"
}

func num2String(num int, format string, ordinal bool) string {
	str := num2words.ConvertAnd(num)
	strSegs := strings.Split(str, " ")
	if ordinal {
		// 如果是序数，那么最后一个单词需要转换
		lastSeg := strSegs[len(strSegs)-1]
		lastSeg2 := getStringOrdinal(lastSeg)
		strSegs[len(strSegs)-1] = lastSeg2
	}

	// 全部单词大写
	for i, seg := range strSegs {
		if lo.Contains(dividerUnit, seg) {
			seg = seg + ","
			strSegs[i] = seg
		}

		if format == "W" {
			strSegs[i] = strings.ToUpper(seg)
		} else if format == "Ww" {
			if seg == "and" {
				continue
			}
			strSegs[i] = strings.Title(seg)
		}
	}

	return strings.Join(strSegs, " ")
}

func GetJsonIndent(v interface{}) string {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err.Error()
	} else {
		return string(out)
	}
}

func FormatInteger(x float64, format string) (string, error) {
	formatOption, err := parseFormat(format)
	if err != nil {
		return "", err
	}

	// 第一步，截断 x 的小数部分
	x = float64(int64(x))

	if lo.Contains([]string{"w", "W", "Ww"}, formatOption.DirectFormat) {
		str := num2String(int(x), formatOption.DirectFormat, formatOption.Ordinal)
		return str, nil
	}

	if lo.Contains([]string{"i", "I"}, formatOption.DirectFormat) {
		roman, err := rom.IntToString(int(x)) // 默认返回大写
		if err != nil {
			return "", err
		}
		if formatOption.DirectFormat == "i" {
			return strings.ToLower(roman), nil
		}
		return roman, nil
	}

	if lo.Contains([]string{"a", "A"}, formatOption.DirectFormat) {
		roman, err := ColumnNumberToName(int(x)) // 默认返回大写
		if err != nil {
			return "", err
		}
		if formatOption.DirectFormat == "a" {
			return strings.ToLower(roman), nil
		}
		return roman, nil
	}

	// 因为当前的 format number 不支持全角，所以需要转换成半角，在最终输出的时候，再转换成全角
	formatHalf := ToHalfWidth(format)
	number, err := FormatNumber(x, formatHalf, dfForInteger)
	if err != nil {
		return "", err
	}
	if formatOption.FullWord {
		return ToFullWidth(number), nil
	}
	return number, nil

}

type layoutOption struct {
	Ordinal      bool
	DirectFormat string
	FullWord     bool // 是否全角，否则是半角
}

var directFormatList = []string{
	"w",  // words
	"W",  // WORDS
	"Ww", // Words
	"i",  // 罗马数组小写
	"I",  // 罗马数组大写
	"a",  // spreadsheet 列名小写
	"A",  // spreadsheet 列名大写
}

var ErrInvalidFormat = fmt.Errorf("invalid format")

func parseFormat(format string) (layoutOption, error) {
	// 参考 js 的 format-integer 的实现， 只接收特定的格式
	// https://www.w3.org/TR/xpath-functions-31/#func-format-integer

	option := layoutOption{}

	// 不能同时存在全角和半角
	formatRunes := []rune(format)
	hasFullWord := false
	for _, r := range formatRunes {
		if len(string(r)) > 1 {
			hasFullWord = true
		}

		if ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') {
			if !lo.Contains(directFormatList, string(r)) && r != 'o' {
				fmt.Println("invalid format", string(r))
				return option, ErrInvalidFormat
			}
		}
	}

	if hasFullWord {
		// 如果有全角，那么所有的数字都必须是全角
		for _, r := range formatRunes {
			if '0' <= r && r <= '9' {
				return option, ErrInvalidFormat
			}
		}
	}

	// num2words
	if strings.HasSuffix(format, ";o") {
		option.Ordinal = true
		format = strings.TrimSuffix(format, ";o")

		if format == "i" || format == "I" {
			return option, errors.New("roman number does not support ordinal")
		}
	}

	if lo.Contains(directFormatList, format) {
		option.DirectFormat = format
	}

	if hasFullWord {
		option.FullWord = true
	}
	return option, nil
}

func ColumnNumberToName(num int) (string, error) {
	if num < 1 {
		return "", fmt.Errorf("incorrect column number %d", num)
	}
	if num > 16384 {
		return "", errors.New("column number exceeds maximum limit")
	}
	var col string
	for num > 0 {
		col = string(rune((num-1)%26+65)) + col
		num = (num - 1) / 26
	}
	return col, nil
}

func NameToColumnNumber(columnName string) (int, error) {
	result := 0
	for _, r := range columnName {
		if r < 'A' || r > 'Z' {
			return 0, fmt.Errorf("invalid column name: %s", columnName)
		}
		result = result*26 + (int(r-'A') + 1)
	}
	return result, nil
}

// 半角转全角
func ToFullWidth(s string) string {
	var fullWidth []rune
	for _, c := range s {
		if unicode.IsPrint(c) && c <= 0x7F && c >= 0x21 {
			c += 0xFEE0
		}
		fullWidth = append(fullWidth, c)
	}
	return string(fullWidth)
}

// 全角转半角
func ToHalfWidth(s string) string {
	var halfWidth []rune
	for _, c := range s {
		if unicode.IsPrint(c) && c >= 0xFF01 && c <= 0xFF5E {
			c -= 0xFEE0
		} else if c == 0x3000 { // 全角空格特殊处理
			c = 0x20
		}
		halfWidth = append(halfWidth, c)
	}
	return string(halfWidth)
}
