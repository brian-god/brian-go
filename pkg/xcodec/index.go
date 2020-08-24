package xcodec

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

/**
 * Copyright (C) @2020 hugo network Co. Ltd
 * 编码和解码工具
 * @author: hugo
 * @version: 1.0
 * @date: 2020/8/3
 * @time: 23:54
 * @description:
 */
// 统一 Err Kind
const (
	// ErrKindUnmarshalConfigErr ...
	ErrKindUnmarshalConfigErr = "unmarshal config err"
	// ErrKindRegisterErr ...
	ErrKindRegisterErr = "register err"
	// ErrKindUriErr ...
	ErrKindUriErr = "uri err"
	// ErrKindRequestErr ...
	ErrKindRequestErr = "request err"
	// ErrKindFlagErr ...
	ErrKindFlagErr = "flag err"
	// ErrKindListenErr ...
	ErrKindListenErr = "listen err"
	// ErrKindAny ...
	ErrKindAny         = "any"
	ErrREQNotMethod    = "请求体中未携带method信息"
	ErrServerException = "调用服务端异常"
)

// 统一模块信息
const (
	// ModConfig ...
	ModConfig = "config"
	// ModApp ...
	ModApp = "app"
	// ModProc ...
	ModProc = "proc"
	// ModGrpcServer ...
	ModGrpcServer = "server.grpc"
	// ModRegistryETCD ...
	ModRegistryETCD = "registry.etcd"
	// ModClientETCD ...
	ModClientETCD = "client.etcd"
	// ModClientGrpc ...
	ModClientGrpc = "client.grpc"
	// ModClientMySQL ...
	ModClientMySQL = "client.mysql"
	// ModRegistryNacos ...
	ModRegistryNacos = "registry.nacos"
)

// UnmarshalByType 反序列化根据类型
//Struct
//	Map
//	Slice
func UnmarshalByType(data interface{}, tp reflect.Type) (reflect.Value, error) {
	// TODO 的先判断是否是指针，指针需要获取原始的类型
	switch tp.Kind() {
	case reflect.Struct:
		//结构体处理
		return UnmarshalStruct(data, tp)
	case reflect.Map:
		//map处理
		return UnmarshalMap(data, tp)
	case reflect.Slice:
		return UnmarshalArray(data, tp)
		//切片处理
	default:
		va, err := BasicUnmarshalByType(data, tp)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(va), nil
	}
	return reflect.Value{}, nil
}

// UnmarshalArray 传输数据对map解析
func UnmarshalMap(data interface{}, dataType reflect.Type) (reflect.Value, error) {
	//返回对应类型的reflect.value
	dataValue := reflect.MakeMap(dataType)
	//如果没有值则直接返回一个空的对象
	if data == nil {
		return dataValue, nil
	}
	//将接收的数据转成切片
	mapData := data.(map[string]interface{})
	dataKind := dataType.Elem().Kind()
	//遍历切片
	for key, v := range mapData {
		switch dataKind {
		case reflect.Struct:
			//结构体处理
			val, err := UnmarshalStruct(v, dataType.Elem())
			if err != nil {
				return reflect.Value{}, err
			}
			dataValue.SetMapIndex(reflect.ValueOf(key), val)
			break
		case reflect.Map:
			//map处理
			val, err := UnmarshalMap(v, dataType.Elem())
			if err != nil {
				return reflect.Value{}, err
			}
			dataValue.SetMapIndex(reflect.ValueOf(key), val)
			break
		case reflect.Slice:
			//切片处理
			val, err := UnmarshalArray(v, dataType.Elem())
			if err != nil {
				return reflect.Value{}, err
			}
			dataValue.SetMapIndex(reflect.ValueOf(key), val)
			break
		default:
			val, err := BasicUnmarshalByType(v, dataType.Elem())
			if err != nil {
				return reflect.Value{}, err
			}
			dataValue.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(val))
			break
		}
	}
	return dataValue, nil
}

// UnmarshalArray 传输数据对数组解析
// TODO 缺少错误处理
func UnmarshalArray(data interface{}, dataType reflect.Type) (reflect.Value, error) {
	//返回对应类型的reflect.value
	dataValue := reflect.New(dataType)
	//如果没有值则直接返回一个空的对象
	if data == nil {
		return dataValue, nil
	}
	//判断是否是指针，只有指针才能进行操作
	if dataValue.Kind() == reflect.Ptr {
		//是否时空的
		if dataValue.IsNil() {
			return reflect.Value{}, errors.New("指针为空")
		}
		// 解引用
		dataValue = dataValue.Elem()
	}
	//将接收的数据转成切片
	mapData := data.([]interface{})
	dataKind := dataType.Elem().Kind()
	//遍历切片
	for _, v := range mapData {
		switch dataKind {
		case reflect.Struct:
			//结构体处理
			val, err := UnmarshalStruct(v, dataType.Elem())
			if err != nil {
				return reflect.Value{}, err
			}
			dataValue = reflect.Append(dataValue, val)
			break
		case reflect.Map:
			//map处理
			val, err := UnmarshalMap(v, dataType.Elem())
			if err != nil {
				return reflect.Value{}, err
			}
			dataValue = reflect.Append(dataValue, val)
			break
		case reflect.Slice:
			//切片处理
			val, err := UnmarshalArray(v, dataType.Elem())
			if err != nil {
				return reflect.Value{}, err
			}
			dataValue = reflect.Append(dataValue, val)
			break
		default:
			val, err := BasicUnmarshalByType(v, dataType.Elem())
			if err != nil {
				return reflect.Value{}, err
			}
			dataValue = reflect.Append(dataValue, reflect.ValueOf(val))
			break
		}
	}
	return dataValue, nil
}

// UnmarshalStruct 传输数据对结构体的解析
func UnmarshalStruct(data interface{}, dataType reflect.Type) (reflect.Value, error) {
	//返回对应类型的reflect.value
	dataValue := reflect.New(dataType)
	//如果没有值则直接返回一个空的对象
	if data == nil {
		return dataValue, nil
	}
	//将接收的数据转成string
	mapData := data.(map[string]interface{})
	//判断是否是指针，只有指针才能进行操作
	if dataValue.Kind() == reflect.Ptr {
		//是否时空的
		if dataValue.IsNil() {
			return reflect.Value{}, errors.New("指针为空")
		}
		// 解引用
		dataValue = dataValue.Elem()
	}
	//获取结构体属性的个数
	fieldNum := dataType.NumField()
	//通过遍历给结构体的属性赋值
	for i := 0; i < fieldNum; i++ {
		field := dataType.Field(i)
		var parValue interface{}
		//获取json表单中的值
		parValue = mapData[field.Name]
		//根据名称获取值信息
		fieldValue := dataValue.Field(i)
		//判断值是否有效。 当值本身非法时，返回 false，例如 reflect Value不包含任何值，值为 nil 等。
		if !fieldValue.IsValid() {
			continue
		}
		if fieldValue.CanInterface() {
			//判断值是否可以被改变
			if fieldValue.CanSet() {
				// TODO 当前只对基本类型处理缺少对结构体中数组和结构体的处理
				switch field.Type.Kind() {
				case reflect.Struct:
					val, err1 := UnmarshalStruct(parValue, field.Type)
					if err1 != nil {
						return reflect.Value{}, err1
					}
					//赋值
					fieldValue.Set(val)
					break
				case reflect.Slice:
					val, err1 := UnmarshalArray(parValue, field.Type)
					if err1 != nil {
						return reflect.Value{}, err1
					}
					//赋值
					fieldValue.Set(val)
					break
				case reflect.Map:
					val, err1 := UnmarshalMap(parValue, dataType.Elem())
					if err1 != nil {
						return reflect.Value{}, err1
					}
					//赋值
					fieldValue.Set(val)
					break
				default:
					//基本本数据类型转换
					val, err1 := BasicUnmarshalByType(parValue, field.Type)
					if err1 != nil {
						return reflect.Value{}, err1
					}
					//赋值
					fieldValue.Set(reflect.ValueOf(val))
					break
				}
			}

		}
	}
	return dataValue, nil
}

//	Bool
//	Int
//	Int8
//	Int16
//	Int32
//	Int64
//	Uint
//	Uint8
//	Uint16
//	Uint32
//	Uint64
//	Uintptr
//	Float32
//	Float64
//	Complex64
//	Complex128
//	String
// BasicUnmarshalByType 基础类型的解码
func BasicUnmarshalByType(data interface{}, tp reflect.Type) (interface{}, error) {
	switch tp.Kind() {
	case reflect.Int:
		var v int
		err := json.Unmarshal([]byte(strconv.Itoa(int(data.(float64)))), &v)
		if err != nil {
			return reflect.Value{}, err
		}
		return v, nil
	case reflect.Int8:
		var v int8
		err := json.Unmarshal([]byte(strconv.Itoa(int(int8(data.(float64))))), &v)
		if err != nil {
			return reflect.Value{}, err
		}
		return v, nil
	case reflect.Int16:
		var v int16
		err := json.Unmarshal([]byte(strconv.Itoa(int(data.(float64)))), &v)
		if err != nil {
			return reflect.Value{}, err
		}
		return v, nil
	case reflect.Int32:
		var v int32
		err := json.Unmarshal([]byte(strconv.Itoa(int(data.(float64)))), &v)
		if err != nil {
			return reflect.Value{}, err
		}
		return v, nil
	case reflect.Int64:
		var v int64
		err := json.Unmarshal([]byte(strconv.Itoa(int(data.(float64)))), &v)
		if err != nil {
			return reflect.Value{}, err
		}
		return v, nil
	case reflect.Uint:
		var v uint
		err := json.Unmarshal([]byte(strconv.Itoa(int(uint(data.(float64))))), &v)
		if err != nil {
			return reflect.Value{}, err
		}
		return v, nil
	case reflect.Uint8:
		var v uint8
		err := json.Unmarshal([]byte(strconv.Itoa(int(uint(data.(float64))))), &v)
		if err != nil {
			return reflect.Value{}, err
		}
		return v, nil
	case reflect.Uint16:
		var v uint16
		err := json.Unmarshal([]byte(strconv.Itoa(int(uint(data.(float64))))), &v)
		if err != nil {
			return reflect.Value{}, err
		}
		return v, nil
	case reflect.Uint32:
		var v uint32
		err := json.Unmarshal([]byte(strconv.Itoa(int(uint(data.(float64))))), &v)
		if err != nil {
			return reflect.Value{}, err
		}
		return v, nil
	case reflect.Uint64:
		var v uint64
		err := json.Unmarshal([]byte(strconv.Itoa(int(uint(data.(float64))))), &v)
		if err != nil {
			return reflect.Value{}, err
		}
		return v, nil
	case reflect.Float32:
		var v float32
		err := json.Unmarshal([]byte(strconv.FormatFloat(data.(float64), 'E', -1, 64)), &v)
		if err != nil {
			return reflect.Value{}, err
		}
		return v, nil
	case reflect.Float64:
		var v float64
		err := json.Unmarshal([]byte(strconv.FormatFloat(data.(float64), 'E', -1, 64)), &v)
		if err != nil {
			return reflect.Value{}, err
		}
		return v, nil
	case reflect.String:
		return data.(string), nil
	case reflect.Bool:
		return data.(bool), nil
	default:
		return nil, errors.New(fmt.Sprintf("无法解析参数，类型为：%s", tp.Kind().String()))
	}
	return nil, nil
}

// Strval 获取变量的字符串值
// 浮点型 3.0将会转换成字符串3, "3"
// 非数值或字符类型的变量将会被转换成JSON格式字符串
func Strval(value interface{}) string {
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return key
}
