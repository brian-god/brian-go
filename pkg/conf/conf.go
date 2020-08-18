package conf

import (
	"github.com/pkg/errors"
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
//配置整个系统的应用
type Configuration struct {
	mu       sync.RWMutex
	override map[string]interface{}
	keyDelim string

	keyMap    *sync.Map
	onChanges []func(*Configuration)

	watchers map[string][]func(*Configuration)
}

//放在默认的配置中
// UnmarshalKey takes a single key and unmarshal it into a Struct with default defaultConfiguration.
func UnmarshalKey(key string, rawVal interface{}, opts ...GetOption) error {
	//配置默认设置
	return nil
	//return defaultConfiguration.UnmarshalKey(key, rawVal, opts...)
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
