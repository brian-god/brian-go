package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/brian-god/brian-go/pkg/utils/xgo"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"net/url"
	"strconv"
	"time"
)

/**
 *
 * Copyright (C) @2020 hugo network Co. Ltd
 * @description
 * @updateRemark
 * @author               hugo
 * @updateUser
 * @createDate           2020/8/20 10:11 上午
 * @updateDate           2020/8/20 10:11 上午
 * @version              1.0
**/

/*
基于http的配置轮询的配置获取
*/
type yaseeDataSource struct {
	lastRevision int64
	enableWatch  bool
	client       *resty.Client
	addr         string
	changed      chan struct{}
	data         string
}

// default client resp struct
type yaseeRes struct {
	Code int        `json:"code"`
	Msg  string     `json:"msg"`
	Data ConfigData `json:"data"`
}

// ConfigData ...
type ConfigData struct {
	Content      string `json:"content"`
	LastRevision int64  `json:"last_revision"`
}

// NewDataSource ...
func NewDataSource(addr string, enableWatch bool) *yaseeDataSource {
	yasee := &yaseeDataSource{
		client:      resty.New(),
		addr:        addr,
		changed:     make(chan struct{}),
		enableWatch: enableWatch,
	}
	if enableWatch {
		xgo.Go(yasee.watch)
	}
	return yasee
}

// ReadConfig ...
func (y *yaseeDataSource) ReadConfig() ([]byte, error) {
	// 检查watch 如果watch为真，走长轮询逻辑
	switch y.enableWatch {
	case true:
		return []byte(y.data), nil
	default:
		content, err := y.getConfigInner(y.addr, y.enableWatch)
		return []byte(content), err
	}
}

// IsConfigChanged ...
func (y *yaseeDataSource) IsConfigChanged() <-chan struct{} {
	return y.changed
}

// Close ...
func (y *yaseeDataSource) Close() error {
	close(y.changed)
	return nil
}

func (y *yaseeDataSource) watch() {
	for {
		resp, err := y.client.R().SetQueryParam("watch", strconv.FormatBool(y.enableWatch)).Get(y.addr)
		// client get err
		if err != nil {
			time.Sleep(time.Second * 1)
			logrus.Error("HttpDataSource", fmt.Sprintf("listenConfig curl err %s", err.Error()))
			continue
		}
		if resp.StatusCode() != 200 {
			time.Sleep(time.Second * 1)
			logrus.Error("HttpDataSource", fmt.Sprintf("listenConfig status err %s", err.Error()))
		}
		var yaseeRes yaseeRes
		if err := json.Unmarshal(resp.Body(), &yaseeRes); err != nil {
			time.Sleep(time.Second * 1)
			logrus.Error("HttpDataSource", fmt.Sprintf("unmarshal err %s", err.Error()))
			continue
		}
		// default code != 200 means not change
		if yaseeRes.Code != 200 {
			time.Sleep(time.Second * 1)
			logrus.Info("HttpDataSource", fmt.Sprintf("code %d", int64(yaseeRes.Code)))
			continue
		}
		select {
		case y.changed <- struct{}{}:
			// record the config change data
			y.data = yaseeRes.Data.Content
			y.lastRevision = yaseeRes.Data.LastRevision
			logrus.Info("HttpDataSource", fmt.Sprintf("change %s", yaseeRes.Data.Content))
		default:
		}
	}
}

// 获取配置
func (y *yaseeDataSource) getConfigInner(addr string, enableWatch bool) (string, error) {
	urlParse, err := url.Parse(addr)
	if err != nil {
		return "", fmt.Errorf("config addr is wrong, err:%v", err.Error())
	}
	appName := urlParse.Query().Get("name")
	appEnv := urlParse.Query().Get("env")
	target := urlParse.Query().Get("target")
	port := urlParse.Query().Get("port")
	commonKey := fmt.Sprintf("%s-%s-%s-%s", appName, appEnv, target, port)
	if commonKey == "" {
		return "", fmt.Errorf("config check key is null")
	}
	content, err := y.getConfig(addr, enableWatch)
	if err != nil {
		content, err = ReadConfigFromFile(commonKey, "config")
		if err != nil {
			return "", errors.New("read config from both server and cache fail")
		}
		return "", err
	}
	WriteConfigToFile(commonKey, "config", content)
	return content, nil
}

func (y *yaseeDataSource) getConfig(addr string, enableWatch bool) (string, error) {
	resp, err := y.client.SetDebug(true).R().SetQueryParam("watch", strconv.FormatBool(enableWatch)).Get(addr)
	if err != nil {
		return "", errors.New("get config err")
	}
	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("get config reply err code:%v", resp.Status())
	}
	configRes := yaseeRes{}
	if err := json.Unmarshal(resp.Body(), &configRes); err != nil {
		return "", fmt.Errorf("unmarshal config err:%v", err.Error())
	}
	if configRes.Code != 200 {
		return "", fmt.Errorf("get config reply err code:%v", resp.Status())
	}
	return configRes.Data.Content, nil
}
