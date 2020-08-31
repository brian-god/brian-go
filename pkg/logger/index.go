package logger

import (
	"github.com/brian-god/brian-go/pkg/xcast"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

/**
 * Copyright (C) @2020 hugo network Co. Ltd
 *
 * @author: hugo
 * @version: 1.0
 * @date: 2020/8/3
 * @time: 23:29
 * @description:
 */

// 模块
func FieldMod(value string) logrus.Fields {
	value = strings.Replace(value, " ", ".", -1)
	return String("mod", value)
}

// FieldMethod ...
func FieldMethod(value string) logrus.Fields {
	return String("method", value)
}

//构建打印结果
func String(key, value string) logrus.Fields {
	return logrus.Fields{key: value}
}

// 依赖的实例名称。以mysql为例，"dsn = "root:juno@tcp(127.0.0.1:3306)/juno?charset=utf8"，addr为 "127.0.0.1:3306"
func FieldAddr(value string) logrus.Fields {
	return String("addr", value)
}

// FieldErrKind ...
func FieldErrKind(value string) logrus.Fields {
	return String("errKind", value)
}

// FieldErr ...
func FieldErr(err error) logrus.Fields {
	return Error(err)
}

// Error is shorthand for the common idiom NamedError("error", err).
func Error(err error) logrus.Fields {
	return NamedError("error", err)
}

// FieldKey ...
func FieldKey(value string) logrus.Fields {
	return String("key", value)
}

func NowTime() logrus.Fields {
	now := time.Now()
	// 24小时制
	value := now.Format("2006-01-02 15:04:05.000 Mon Jan")
	return String("time", value)
}

// NamedError constructs a field that lazily stores err.Error() under the
// provided key. Errors which also implement fmt.Formatter (like those produced
// by github.com/pkg/errors) will also have their verbose representation stored
// under key+"Verbose". If passed a nil error, the field is a no-op.
//
// For the common case in which the key is simply "error", the Error function
// is shorter and less repetitive.
func NamedError(key string, err error) logrus.Fields {
	if err == nil {
		return Skip()
	}
	return logrus.Fields{"key": key, "错误": err}
}

// Skip constructs a no-op field, which is often useful when handling invalid
// inputs in other Field constructors.
func Skip() logrus.Fields {
	return logrus.Fields{"error": "未知异常"}
}
func Any(key string, value interface{}) logrus.Fields {
	return String("key", xcast.ToString(value))
}
