package utils

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
)

func GetJsonIndent(data interface{}) string {
	switch data.(type) {
	case reflect.Value:
		// check if data is zero
		if data == reflect.ValueOf(nil) {
			return "null"
		}
		data = data.(reflect.Value).Interface()
	case []reflect.Value:
		// check if data is zero
		// convert []reflect.Value to []interface{}
		values := data.([]reflect.Value)
		data = make([]interface{}, len(values))
		for i, v := range values {
			if v == reflect.ValueOf(nil) {
				continue
			}
			if v.IsZero() || !v.CanInterface() {
				continue
			}
			data.([]interface{})[i] = v.Interface()
		}
	}
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		Log("GetJsonIndent error:", err)
		return ""
	}
	return string(out)
}

func Log(a ...any) (n int, err error) {
	// 获取调用 Log 函数的行号
	_, file, line, _ := runtime.Caller(1)
	// 获取文件名
	file = "> " + filepath.Base(file)
	file = file + ":" + fmt.Sprint(line)
	a = append([]any{file}, a...)

	return fmt.Println(a...)
}
