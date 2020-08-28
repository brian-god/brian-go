package xstring

import (
	"reflect"
	"runtime"
)

/**
 *
 * Copyright (C) @2020 hugo network Co. Ltd
 * @description
 * @updateRemark
 * @author               hugo
 * @updateUser
 * @createDate           2020/8/28 4:50 下午
 * @updateDate           2020/8/28 4:50 下午
 * @version              1.0
**/

// FunctionName ...
func FunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// ObjectName ...
func ObjectName(i interface{}) string {
	typ := reflect.TypeOf(i)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	return typ.PkgPath() + "." + typ.Name()
}

// CallerName ...
func CallerName(skip int) string {
	pc, _, _, _ := runtime.Caller(skip)
	return runtime.FuncForPC(pc).Name()
}
