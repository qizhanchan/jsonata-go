package jxpath

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"
)

// 参考 https://www.w3.org/TR/xpath-functions-31/#rules-for-datetime-formatting 实现

type DateIndex string

const (
	Year              DateIndex = "Y"
	Month             DateIndex = "M"
	Day               DateIndex = "D"
	DayOfYear         DateIndex = "d"
	DayOfWeek         DateIndex = "F"
	WeekOfYear        DateIndex = "W"
	WeekOfMonth       DateIndex = "w"
	Hour              DateIndex = "H"
	HourHalf          DateIndex = "h"
	AmPm              DateIndex = "P"
	Minute            DateIndex = "m"
	Second            DateIndex = "s"
	FractionalSeconds DateIndex = "f"
	Timezone          DateIndex = "Z"
)

const (
	MinWidthSep = ','
	MaxWidthSep = '-'
)

type dateSegmentOption struct {
	Index DateIndex

	toWords           bool
	toWordsFormat     string
	toNumbers         bool
	toNumbersFormat   string
	toNumbersMinWidth int
	toNumbersMaxWidth int
}

type patternIndex int

const (
	numIndex  patternIndex = iota + 1
	textIndex              // 文本，非数字，也非符号
)

type numSeg struct {
	indexType patternIndex
	data      string
}

func maxPadding(str, format string) string {
	if len(str) >= len(format) {
		return str
	}
	paddingZeroLen := len(format) - len(str)
	paddingZero := strings.Repeat("0", paddingZeroLen)
	return paddingZero + str
}

func cutString(minWidth, maxWidth int, strCopy, strOrigin string) string {
	if maxWidth == 0 {
		return "0"
	}

	if minWidth < 0 && maxWidth < 0 {
		return strCopy
	}

	// 只设置了最小长度
	if minWidth >= 0 && maxWidth < 0 {
		if minWidth > len(strCopy) {
			// 需要重置到原来的值，再计算最小长度
			strCopy = strOrigin
			if minWidth > len(strCopy) {
				paddingZeroLen := minWidth - len(strCopy)
				paddingZero := strings.Repeat("0", paddingZeroLen)
				return paddingZero + strCopy
			} else {
				// minWidth 大于原来的长度，是不生效的, 直接返回原来的值
				return strCopy
			}
		}
	}
	if maxWidth >= 0 {
		if maxWidth > len(strCopy) {
			paddingZeroLen := maxWidth - len(strCopy)
			paddingZero := strings.Repeat("0", paddingZeroLen)
			return paddingZero + strCopy
		} else {
			// 如果最大长度小于原来的长度，则切片
			return strCopy[len(strCopy)-maxWidth:]
		}
	}
}

func getNamedStr(str, format string) string {
	switch format {
	case "n":
		return strings.ToLower(str)
	case "N":
		return strings.ToUpper(str)
	case "Nn":
		return strings.ToUpper(str[:1]) + strings.ToLower(str[1:])

	}
	return str
}

// 这里的实现和 float 的实现不一样，逗号分隔符的实现不一样，float 的会作为一个整体，这里会作为一个分隔符
func formatNumber(num int, format string, index DateIndex) (string, error) {

	var (
		minWidth = -1
		maxWidth = -1
	)
	formatSegs := strings.Split(format, string(MinWidthSep))
	if len(formatSegs) > 2 {
		return "", errors.New("format error: has more than 1 min width separator")
	}

	// 提取最小宽度和最大宽度
	if len(formatSegs) == 2 {
		lengthIndex := formatSegs[1]
		lengthSegs := strings.Split(lengthIndex, string(MaxWidthSep))
		if len(lengthSegs) > 2 {
			return "", errors.New("format error: has more than 1 max width separator")
		}

		if length, err := strconv.Atoi(lengthSegs[0]); err == nil {
			minWidth = length
		}
		if length, err := strconv.Atoi(lengthSegs[1]); err == nil {
			maxWidth = length
		}

		if maxWidth < minWidth {
			return "", errors.New("format error: max width less than min width")
		}
	}

	sourceStr := strconv.Itoa(num)
	// 把连续的 ## 替换成 #
	if strings.Contains(format, "##") {
		format = strings.ReplaceAll(format, "##", "#0")
	}

	// 如果有 # 符号，则需要特殊处理一下
	// 当年份是 2018 的时候
	// #12 > 12
	// 12#12 > 12
	// #012 > 2018
	// 1#234567 > 0002018
	// 对应的逻辑是：# 后面数字长度是 1,2 的时候，输出 2 位数字，尝试是 3,4 的时候，输出 4 位数字

	// 假设我们有一个数字格式：00l0m99，需要分割为
	// 数字列表：[00, 0, 99]
	// 分隔符列表：[l, m]

	if strings.HasPrefix(format, "#") {
		format2 := format[1:]
		if len(format2) == 3 {
			format2 += "0"
		}

		sourceStrCopy := sourceStr
		if len(format2) < len(sourceStr) {
			sourceStrCopy = sourceStr[len(sourceStr)-len(format2):]
		}

		return cutString(minWidth, maxWidth, sourceStrCopy, sourceStr), nil
	}

	// 只有 day of week 和 月份才支持 name
	if len(format) > 0 && strings.ToLower(string(format[0])) == "n" {
		if lo.Contains([]DateIndex{DayOfWeek, Month}, index) {
			if index == DayOfWeek {
				sourceStr = time.Weekday(num).String()
				sourceStr = getNamedStr(sourceStr, format)
				sourceStr = cutString(minWidth, maxWidth, sourceStr, sourceStr)
				return sourceStr, nil
			}

			if index == Month {
				sourceStr = time.Month(num).String()
				sourceStr = getNamedStr(sourceStr, format)
				sourceStr = cutString(minWidth, maxWidth, sourceStr, sourceStr)
				return sourceStr, nil
			}
		}
		return "", errors.New("format error: name format only support day of week and month")
	}

	// format 是纯数字场景

	formatLen := len(format)
	switch index {
	case Year:
		if formatLen < 2 {
			return cutString(minWidth, maxWidth, sourceStr, sourceStr), nil
		} else if formatLen < 5 {
			sourceStrCopy := sourceStr[len(sourceStr)-formatLen-2:]
			return cutString(minWidth, maxWidth, sourceStrCopy, sourceStr), nil
		} else {
			return "", errors.New("format error: year format length is too long")
		}

	case Month, Day, WeekOfMonth:
		if formatLen == 1 {
			return sourceStr, nil
		} else {

		}

	case DayOfYear, WeekOfYear:
		return sourceStr, nil

	}
}

func eatNumber(str string, index int) (int, string) {
	numStr := ""
	for i := index; i < len(str); i++ {
		if '0' <= str[i] && str[i] <= '9' {
			numStr += string(str[i])
		} else {
			break
		}
	}
	return index + len(numStr), numStr
}

// Y	year (absolute value)	1
// M	month in year	1
// D	day in month	1
// d	day in year	1
// F	day of week	n
// W	week in year	1
// w	week in month	1
// H	hour in day (24 hours)	1
// h	hour in half-day (12 hours)	1
// P	am/pm marker	n
// m	minute in hour	01
// s	second in minute	01
// f	fractional seconds	1
// Z	timezone	01:01

func segmentToString(o dateSegmentOption) string {
	switch o.Index {
	case Year:

	}
}
