// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package jlib

import (
	"errors"
	"reflect"

	"github.com/blues/jsonata-go/jtypes"
	"github.com/blues/jsonata-go/utils"
)

// Boolean (golint)
func Boolean(v reflect.Value) bool {
	if v == reflect.ValueOf(nil) {
		utils.Log("jlib.Boolean undefined")
	}

	v = jtypes.Resolve(v)

	if b, ok := jtypes.AsBool(v); ok {
		return b
	}

	if s, ok := jtypes.AsString(v); ok {
		return s != ""
	}

	if n, ok := jtypes.AsNumber(v); ok {
		return n != 0
	}

	if jtypes.IsArray(v) {
		for i := 0; i < v.Len(); i++ {
			if Boolean(v.Index(i)) {
				return true
			}
		}
		return false
	}

	if jtypes.IsMap(v) {
		return v.Len() > 0
	}

	return false
}

// Not (golint)
func Not(v reflect.Value) BoolEx {
	// check if v is jtypes.NoMatched
	if jtypes.IsEqual(v, jtypes.NoMatchedCtx) {
		utils.Log("get NoMatched 1")
		return BoolEx{Data: false, Ctx: jtypes.NoMatchedCtx}
	}

	res := !Boolean(v)
	boolEx := BoolEx{Data: res}
	return boolEx
}

// Exists (golint)
func Exists(v reflect.Value) bool {
	return v.IsValid()
}

type BoolEx struct {
	Data bool
	Ctx  jtypes.TransCtx
}

// Assert (golint)
func Assert(v1, v2 reflect.Value) interface{} {
	utils.Log("Assert", utils.GetJsonIndent(v1), utils.GetJsonIndent(v2))
	if b, ok := jtypes.AsBool(v1); ok {
		if b {
			str, ok := jtypes.AsString(v2)
			if !ok {
				return errors.New("assert 2nd argument is not string")
			}
			utils.Log("return err string")
			return errors.New(str)
		} else {
			return nil
		}
	} else {
		panic("Assert error, need bool as 1st argument")
	}
	return nil
}
