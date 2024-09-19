package jxpath

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"
)

var (
	tplOfDate = "\\[[\\w-,]+\\]"
	regOfDate = regexp.MustCompile(tplOfDate)
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

	return strCopy
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

	ordinal := false
	if strings.HasSuffix(format, "o") {
		ordinal = true
		format = format[:len(format)-1]
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
		if lo.Contains([]DateIndex{DayOfWeek, Month, AmPm}, index) {
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

			if index == AmPm {
				sourceStr := "Am"
				// 对于 PM AM 的判断，使用的是 hour 的值
				if num > 12 {
					sourceStr = "Pm"
				}
				sourceStr = getNamedStr(sourceStr, format)
				sourceStr = cutString(minWidth, maxWidth, sourceStr, sourceStr)
				return sourceStr, nil
			}
		}
		return "", errors.New("format error: name format only support day of week and month")
	}

	// format 是纯数字场景

	num2, err := getDateNum(num, sourceStr, format, index)
	if err != nil {
		return "", err
	}

	// 进行最小宽度和最大宽度的处理
	num2 = cutString(minWidth, maxWidth, num2, sourceStr)

	if ordinal {
		num2 = num2Ordinal(num2)
	}
	return num2, nil
}

// 分钟和秒，默认是 00
var zeroNum = map[DateIndex]string{
	Minute: "00",
	Second: "00",
}

func getDateNum(num int, sourceStr, format string, index DateIndex) (res string, err error) {
	if num == 0 {
		if zeroNumStr, ok := zeroNum[index]; ok {
			sourceStr = zeroNumStr
		}
	}

	formatLen := len(format)
	switch index {
	case Year:

		if formatLen < 2 {
			return sourceStr, nil
		} else if formatLen < 5 {
			sourceStrCopy := sourceStr[len(sourceStr)-formatLen-2:]
			return sourceStrCopy, nil
		} else {
			return "", errors.New("format error: year format length is too long")
		}

		// 这三个变量都是
	case Month, Day, WeekOfMonth, Hour, HourHalf, Minute, Second, FractionalSeconds:
		if formatLen < 2 {
			return sourceStr, nil
		} else if formatLen < 4 {
			if len(sourceStr) < formatLen {
				// padding 0
				paddingZeroLen := formatLen - len(sourceStr)
				paddingZero := strings.Repeat("0", paddingZeroLen)
				return paddingZero + sourceStr, nil
			}
			return sourceStr, nil
		} else {
			return "", errors.New("format error: month, day, week of month format length is too long")
		}

		// 不受格式符号影响
	case DayOfYear, WeekOfYear:
		return sourceStr, nil

	case Timezone:
		format = strings.TrimSuffix(format, "t")
		return FormatInteger(float64(num), format)
	}

	return "", fmt.Errorf("format error: unknown date index:%v", index)

}

// 把一个数字转换成顺序模式
// 1 -> 1st
// 2 -> 2nd
// 20 -> 20th
// 01 -> 1st
func num2Ordinal(num string) string {
	lastDigit := num[len(num)-1]
	switch lastDigit {
	case '1':
		return num + "st"
	case '2':
		return num + "nd"
	case '3':
		return num + "rd"
	default:
		return num + "th"
	}
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

var prefixMap = map[string]DateIndex{
	"Y": Year,
	"M": Month,
	"D": Day,
	"d": DayOfYear,
	"F": DayOfWeek,
	"W": WeekOfYear,
	"w": WeekOfMonth,
	"H": Hour,
	"h": HourHalf,
	"P": AmPm,
	"m": Minute,
	"s": Second,
	"f": FractionalSeconds,
	"Z": Timezone,
}

func segmentToString(t time.Time, format string) (string, error) {
	if format == "" {
		return "", nil
	}
	prefix := format[0]
	index, ok := prefixMap[string(prefix)]
	if !ok {
		return "", fmt.Errorf("format error: unknown date prefix:%v", string(prefix))
	}
	format = format[1:]

	fmt.Println("index", index)

	num := 0

	switch index {
	case Year:
		num = t.Year()
	case Month:
		num = int(t.Month())
	case Day:
		num = t.Day()
	case DayOfYear:
		num = t.YearDay()
	case DayOfWeek:
		num = int(t.Weekday())
	case WeekOfYear:
		_, week := t.ISOWeek()
		num = week
	case WeekOfMonth:
		_, week := t.ISOWeek()
		num = week
	case Hour:
		num = t.Hour()
	case HourHalf:
		num = t.Hour() % 12
	case AmPm:
		num = t.Hour()
	case Minute:
		num = t.Minute()
	case Second:
		num = t.Second()
	case FractionalSeconds:
		num = t.Nanosecond() / 1000000
	case Timezone:
		_, offset := t.Zone()
		num = offset/3600*100 + offset%3600/60
	}

	// fmt.Println("num", num, "index", index)
	res, err := formatNumber(num, format, index)
	if err != nil {
		return "", err
	}
	if index == Timezone {
		_, offset := t.Zone()
		if offset > 0 {
			res = "+" + res
		}
		// else {
		// 	// res = "-" + res
		// }
	}
	return res, nil
}

func FromMills(unixMills int, format string, timezone string) (string, error) {
	// 替换换行符
	format = strings.ReplaceAll(format, "\n", "")

	// locate timezone
	// 解析 "-0400" 时区
	offsetSeconds := 0
	if timezone != "" {
		if len(timezone) != 5 {
			return "", errors.New("timezone format error")
		}

		if timezone[0] != '+' && timezone[0] != '-' {
			return "", errors.New("timezone format error")
		}

		offsetHourStr := timezone[1:3]
		offsetMinuteStr := timezone[3:]
		offsetHour, err := strconv.Atoi(offsetHourStr)
		if err != nil {
			return "", errors.New("timezone format error")
		}
		offsetMinute, err := strconv.Atoi(offsetMinuteStr)
		if err != nil {
			return "", errors.New("timezone format error")
		}
		offsetSeconds = offsetHour*3600 + offsetMinute*60
		if timezone[0] == '-' {
			offsetSeconds = -offsetSeconds
		}
	}

	loc := time.FixedZone("my loc", offsetSeconds)

	// calculate time with unixMills and timezone
	t := time.Unix(int64(unixMills/1000), 0).In(loc)
	fmt.Println("t", t)

	// find all date format by regex, and replace them
	var err2 error
	output := regOfDate.ReplaceAllStringFunc(format, func(match string) string {
		// 在这里对匹配到的字符串进行处理
		formatInner := strings.TrimPrefix(match, "[")
		formatInner = strings.TrimSuffix(formatInner, "]")
		res, err := segmentToString(t, formatInner)
		fmt.Println("res", res)
		if err != nil {
			err2 = err
			return ""
		}

		return res
	})

	if err2 != nil {
		fmt.Println("err2", err2)
		return "", err2
	}

	return output, nil
}
