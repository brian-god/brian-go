package conf

import (
	"fmt"
	"github.com/brian-god/brian-go/pkg/xcast"
	"github.com/brian-god/brian-go/pkg/xcodec"
	"github.com/brian-god/brian-go/pkg/xmap"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"sync"
)

/**
 * Copyright (C) @2020 hugo network Co. Ltd
 *
 * @author: hugo
 * @version: 1.0
 * @date: 2020/8/2
 * @time: 12:27
 * @description:
 */

// Configuration provides configuration for application.
var configType = "properties"

//配置整个系统的应用
type Configuration struct {
	mu       sync.RWMutex
	override map[string]interface{}
	keyDelim string

	keyMap    *sync.Map
	onChanges []func(*Configuration)

	watchers map[string][]func(*Configuration)
}

//获取配置文件类型
func GetConfigType() string {
	return configType
}

//设置配置文件类型
func SetConfigType(v string) {
	configType = v
}

//放在默认的配置中
// UnmarshalKey takes a single key and unmarshal it into a Struct with default defaultConfiguration.
func UnmarshalKey(key string, rawVal interface{}, opts ...GetOption) error {
	//配置默认设置
	return defaultConfiguration.UnmarshalKey(key, rawVal, opts...)
}

//解析配置
func Unmarshal(rawVal interface{}) error {
	//配置默认设置
	return defaultConfiguration.Unmarshal(rawVal)
}

// UnmarshalKey takes a single key and unmarshal it into a Struct.
func (c *Configuration) Unmarshal(rawVal interface{}) error {
	//返回对应的类型
	dataType := reflect.TypeOf(rawVal)
	//返回对应类型的reflect.value
	dataValue := reflect.ValueOf(rawVal)
	//判断是否是指针，只有指针才能进行操作
	if dataValue.Kind() == reflect.Ptr {
		//是否时空的
		if dataValue.IsNil() {
			return errors.New("读取配置文件传入的必须是指针")
		}
		// 解引用
		dataValue = dataValue.Elem()
		dataType = dataType.Elem()
	}
	//获取结构体属性的个数
	fieldNum := dataValue.NumField()
	//从config中获取属性的tag
	tagName := configType
	//通过遍历给结构体的属性赋值
	for i := 0; i < fieldNum; i++ {
		field := dataType.Field(i)
		//获取结构体的tag
		tag := field.Tag.Get(tagName)
		//根据名称获取值信息
		fieldValue := dataValue.Field(i)
		if tag == "" {
			c.mu.RLock()
		}
		//获取配置的值
		value := c.Get(tag)
		if value == nil {
			continue
		}
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
					val, err1 := xcodec.UnmarshalStruct(value, field.Type)
					if err1 != nil {
						return err1
					}
					//赋值
					fieldValue.Set(val)
					break
				case reflect.Slice:
					val, err1 := xcodec.UnmarshalArray(value, field.Type)
					if err1 != nil {
						return err1
					}
					//赋值
					fieldValue.Set(val)
					break
				case reflect.Map:
					val, err1 := xcodec.UnmarshalMap(value, dataType.Elem())
					if err1 != nil {
						return err1
					}
					//赋值
					fieldValue.Set(val)
					break
				default:
					//基本本数据类型转换
					val, err1 := xcodec.BasicUnmarshalByType1(value, field.Type)
					if err1 != nil {
						return err1
					}
					//赋值
					fieldValue.Set(reflect.ValueOf(val))
					break
				}
			}

		}
	}
	return nil
}

const (
	defaultKeyDelim = "."
)

// ErrInvalidKey ...
var ErrInvalidKey = errors.New("invalid key, maybe not exist in config")

// New constructs a new Configuration with provider.
func New() *Configuration {
	return &Configuration{
		override:  make(map[string]interface{}),
		keyDelim:  defaultKeyDelim,
		keyMap:    &sync.Map{},
		onChanges: make([]func(*Configuration), 0),
		watchers:  make(map[string][]func(*Configuration)),
	}
}

// LoadFromDataSource ...
func (c *Configuration) LoadFromDataSource(ds DataSource, unmarshaller Unmarshaller) error {
	content, err := ds.ReadConfig()
	if err != nil {
		return err
	}

	if err := c.Load(content, unmarshaller); err != nil {
		return err
	}

	go func() {
		for range ds.IsConfigChanged() {
			if content, err := ds.ReadConfig(); err == nil {
				_ = c.Load(content, unmarshaller)
				for _, change := range c.onChanges {
					change(c)
				}
			}
		}
	}()

	return nil
}

// Load ...
func (c *Configuration) Load(content []byte, unmarshaller Unmarshaller) error {
	configuration := make(map[string]interface{})
	if err := unmarshaller(content, &configuration); err != nil {
		return err
	}
	return c.apply(configuration)
}

// Load loads configuration from provided data source.
func (c *Configuration) LoadFromReader(reader io.Reader, unmarshaller Unmarshaller) error {
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	return c.Load(content, unmarshaller)
}

// Set ...
func (c *Configuration) Set(key string, val interface{}) error {
	paths := strings.Split(key, c.keyDelim)
	lastKey := paths[len(paths)-1]
	m := deepSearch(c.override, paths[:len(paths)-1])
	m[lastKey] = val
	return c.apply(m)
	// c.keyMap.Store(key, val)
}

//应用配置
func (c *Configuration) apply(conf map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var changes = make(map[string]interface{})

	xmap.MergeStringMap(c.override, conf)
	for k, v := range c.traverse(c.keyDelim) {
		orig, ok := c.keyMap.Load(k)
		if ok && !reflect.DeepEqual(orig, v) {
			changes[k] = v
		}
		c.keyMap.Store(k, v)
	}

	if len(changes) > 0 {
		c.notifyChanges(changes)
	}

	return nil
}

// UnmarshalKey takes a single key and unmarshal it into a Struct.
func (c *Configuration) UnmarshalKey(key string, rawVal interface{}, opts ...GetOption) error {
	var options = defaultGetOptions
	for _, opt := range opts {
		opt(&options)
	}

	config := mapstructure.DecoderConfig{
		DecodeHook: mapstructure.StringToTimeDurationHookFunc(),
		Result:     rawVal,
		TagName:    options.TagName,
	}
	decoder, err := mapstructure.NewDecoder(&config)
	if err != nil {
		return err
	}
	if key == "" {
		c.mu.RLock()
		defer c.mu.RUnlock()
		return decoder.Decode(c.override)
	}

	value := c.Get(key)
	if value == nil {
		return errors.Wrap(ErrInvalidKey, key)
	}

	return decoder.Decode(value)
}

//通知改变的进行改变
func (c *Configuration) notifyChanges(changes map[string]interface{}) {
	var changedWatchPrefixMap = map[string]struct{}{}

	for watchPrefix := range c.watchers {
		for key := range changes {
			// 前缀匹配即可
			// todo 可能产生错误匹配
			if strings.HasPrefix(key, watchPrefix) {
				changedWatchPrefixMap[watchPrefix] = struct{}{}
			}
		}
	}

	for changedWatchPrefix := range changedWatchPrefixMap {
		for _, handle := range c.watchers[changedWatchPrefix] {
			go handle(c)
		}
	}
}

//深度搜索
func deepSearch(m map[string]interface{}, path []string) map[string]interface{} {
	for _, k := range path {
		m2, ok := m[k]
		if !ok {
			m3 := make(map[string]interface{})
			m[k] = m3
			m = m3
			continue
		}
		m3, ok := m2.(map[string]interface{})
		if !ok {
			m3 = make(map[string]interface{})
			m[k] = m3
		}
		m = m3
	}
	return m
}

// Get returns the value associated with the key
func (c *Configuration) Get(key string) interface{} {
	return c.find(key)
}

//查找
func (c *Configuration) find(key string) interface{} {
	//直接读取
	dd, ok := c.keyMap.Load(key)
	if ok {
		return dd
	}

	paths := strings.Split(key, c.keyDelim)
	c.mu.RLock()
	defer c.mu.RUnlock()
	m := xmap.DeepSearchInMap(c.override, paths[:len(paths)-1]...)
	dd = m[paths[len(paths)-1]]
	c.keyMap.Store(key, dd)
	return dd
}

//遍历
func (c *Configuration) traverse(sep string) map[string]interface{} {
	data := make(map[string]interface{})
	lookup("", c.override, data, sep)
	return data
}

func lookup(prefix string, target map[string]interface{}, data map[string]interface{}, sep string) {
	for k, v := range target {
		pp := fmt.Sprintf("%s%s%s", prefix, sep, k)
		if prefix == "" {
			pp = k
		}
		if dd, err := xcast.ToStringMapE(v); err == nil {
			lookup(pp, dd, data, sep)
		} else {
			data[pp] = v
		}
	}
}
