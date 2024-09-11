package jxpath

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	rom "github.com/brandenc40/romannumeral"
	"github.com/divan/num2words"
	"github.com/samber/lo"
)

// func formatDigits(format string) (string, error) {
// 	formatRuns := []rune(format)
// 	lastRune := formatRuns[len(formatRuns)-1]
// 	fullWord := false
// 	if ('0' <= lastRune && lastRune <= '9') && ('０' < lastRune && lastRune < '９') {
//
// 	} else {
// 		return "", fmt.Errorf("last char is not a digit")
// 	}
//
// }

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

func num2String(num int, format string, ordinal bool) string {
	str := num2words.ConvertAnd(num)
	strSegs := strings.Split(str, " ")
	segCount := len(strSegs)
	if ordinal {
		if num == 0 {
			strSegs[0] = ordinalMap[0]
			return strSegs[0]
		}

		if num%100 > 0 {
			mod := num % 100
			if ordinalNum, ok := ordinalMap[mod]; ok {
				strSegs[segCount-1] = ordinalNum
			} else if mod%10 == 0 {
				// 不是 0， 也不是 10， 意味着是 20, 30, 40, 50, 60, 70, 80, 90
				// 需要把最后一位的 y 替换成 ieth
				strSegs[segCount-1] = strings.TrimSuffix(strSegs[segCount-1], "y") + "ieth"
			} else {
				// 处理 21 - 99 并且最后一位不是 0 的情况
				modTen := mod % 10
				if ordinalNum2, ok2 := ordinalMap[modTen]; ok2 {
					// 最后两个单词需要用中横线连接
					strSegs[segCount-1] = num2words.ConvertAnd(mod-modTen) + "-" + ordinalNum2
				}
			}
		} else {
			strSegs[segCount-1] += "th"
		}
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

	if len(strSegs) > 1 {
		return strings.Join(strSegs, " ")
	} else {
		return strSegs[0]
	}
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
	formatOption, err := validateFormat(format)
	if err != nil {
		return "", err
	}

	// 第一步，截断 x 的小数部分
	x = float64(int64(x))

	if formatOption.OnlyNumber {
		fmt.Println("formatOption.OnlyNumber")
		return FormatNumber(x, format, dfForInteger)
	}

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

	return "", nil
}

type layoutOption struct {
	OnlyNumber   bool
	Ordinal      bool
	DirectFormat string
	Full         bool // 是否全角，否则是半角
}

var directFormatList = []string{
	"w",  // words
	"W",  // WORDS
	"Ww", // Words
	"i",  // 罗马数组小写
	"I",  // 罗马数组大写
}

var ErrInvalidFormat = fmt.Errorf("invalid format")

func validateFormat(format string) (layoutOption, error) {
	// 参考 js 的 format-integer 的实现， 只接收特定的格式
	// https://www.w3.org/TR/xpath-functions-31/#func-format-integer

	option := layoutOption{}
	if strings.Contains(format, "#") {
		option.OnlyNumber = true
		return option, nil
	}

	// 不能同时存在全角和半角
	formatRuns := []rune(format)
	hasFullWord := false
	hasHalfWord := false
	for _, r := range formatRuns {
		if len(string(r)) > 1 {
			hasFullWord = true
			break
		} else if len(string(r)) == 1 {
			hasHalfWord = true
		}
	}

	if hasFullWord && hasHalfWord {
		return option, ErrInvalidFormat
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

	return option, nil
}