package jxpath

import (
	"fmt"
	"strconv"
	"strings"

	rom "github.com/brandenc40/romannumeral"
	"github.com/samber/lo"
)

func init() {
	for k, v := range ordinal2cardinal {
		cardinal2ordinal[v] = k
	}
}

// 定义基本数字的映射
var smallNumbers = map[string]int{
	"zero":      0,
	"one":       1,
	"two":       2,
	"three":     3,
	"four":      4,
	"five":      5,
	"six":       6,
	"seven":     7,
	"eight":     8,
	"nine":      9,
	"ten":       10,
	"eleven":    11,
	"twelve":    12,
	"thirteen":  13,
	"fourteen":  14,
	"fifteen":   15,
	"sixteen":   16,
	"seventeen": 17,
	"eighteen":  18,
	"nineteen":  19,
	"twenty":    20,
	"thirty":    30,
	"forty":     40,
	"fifty":     50,
	"sixty":     60,
	"seventy":   70,
	"eighty":    80,
	"ninety":    90,
}

var ordinal2cardinal = map[string]string{
	"zeroth":      "zero",
	"first":       "one",
	"second":      "two",
	"third":       "three",
	"fourth":      "four",
	"fifth":       "five",
	"sixth":       "six",
	"seventh":     "seven",
	"eighth":      "eight",
	"ninth":       "nine",
	"tenth":       "ten",
	"eleventh":    "eleven",
	"twelfth":     "twelve",
	"thirteenth":  "thirteen",
	"fourteenth":  "fourteen",
	"fifteenth":   "fifteen",
	"sixteenth":   "sixteen",
	"seventeenth": "seventeen",
	"eighteenth":  "eighteen",
	"nineteenth":  "nineteen",
	"twentieth":   "twenty",
	"thirtieth":   "thirty",
	"fortieth":    "forty",
	"fiftieth":    "fifty",
	"sixtieth":    "sixty",
	"seventieth":  "seventy",
	"eightieth":   "eighty",
	"ninetieth":   "ninety",
	"hundredth":   "hundred",
	"thousandth":  "thousand",
	"millionth":   "million",
	"billionth":   "billion",
	"trillionth":  "trillion",
}

var cardinal2ordinal = map[string]string{}

var bigNumbers = map[string]int{
	"hundred":  1e2,
	"thousand": 1e3,
	"million":  1e6,
	"billion":  1e9,
	"trillion": 1e12,
}

func wordsToNumber(words string, ordinal bool) int {
	total := 0
	current := 0
	words = strings.Replace(words, "-", " ", -1)
	words = strings.Replace(words, ",", "", -1)
	words = strings.Replace(words, " and ", " ", -1)

	// 将输入的字符串转换为小写，并以空格分割成单词
	words = strings.ToLower(words)
	wordList := strings.Fields(words)

	if ordinal {
		lastWord := wordList[len(wordList)-1]
		wordList[len(wordList)-1] = ordinal2cardinal[lastWord]
	}

	for _, word := range wordList {
		// 检查单词是否是基本数字
		if value, found := smallNumbers[word]; found {

			current += value
			fmt.Println("small value:", value, current)
		} else if value, found := bigNumbers[word]; found {
			current *= value
			total += current
			current = 0
			fmt.Println("big value:", value, current, total)
		}
	}

	return total + current
}

func ParseInteger(num, format string) (int, error) {
	formatOption, err := parseFormat(format)
	if err != nil {
		return 0, err
	}

	// num words
	if lo.Contains([]string{"w", "W", "Ww"}, formatOption.DirectFormat) {
		numLower := strings.ToLower(num)
		numInteger := wordsToNumber(numLower, formatOption.Ordinal)
		return numInteger, nil
	}

	// roman numerals
	if lo.Contains([]string{"i", "I"}, formatOption.DirectFormat) {
		numUpper := strings.ToUpper(num)
		numRoman, err := rom.StringToInt(numUpper) // 默认返回大写
		if err != nil {
			return 0, err
		}
		return numRoman, nil
	}

	// columns
	if lo.Contains([]string{"a", "A"}, formatOption.DirectFormat) {
		numUpper := strings.ToUpper(num)
		numColumn, err := NameToColumnNumber(numUpper) // 默认返回大写
		if err != nil {
			return 0, err
		}
		return numColumn, nil
	}

	// arabic numerals
	num = ToHalfWidth(num)
	arabicNum := ""
	for _, r := range num {
		if r >= '0' && r <= '9' {
			arabicNum += string(r)
		}
	}
	numInt, err := strconv.Atoi(arabicNum)
	if err != nil {
		return 0, err
	}

	return numInt, nil
}
