package file

import (
	"encoding/json"
	"github.com/brian-god/brian-go/pkg/conf"
	"github.com/brian-god/brian-go/pkg/logger"
	"github.com/brian-god/brian-go/pkg/utils/xgo"
	"github.com/brian-god/brian-go/pkg/xfile"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

/**
 *
 * Copyright (C) @2020 hugo network Co. Ltd
 * @description
 * @updateRemark
 * @author               hugo
 * @updateUser
 * @createDate           2020/8/20 11:00 上午
 * @updateDate           2020/8/20 11:00 上午
 * @version              1.0
**/

// fileDataSource file provider.
type fileDataSource struct {
	path        string
	dir         string
	enableWatch bool
	changed     chan struct{}
}

// NewDataSource returns new fileDataSource.
func NewDataSource(path string, watch bool) *fileDataSource {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		logrus.Panicf("new datasource err", err.Error())
	}
	//检查并获取文件夹
	dir := xfile.CheckAndGetParentDir(absolutePath)
	ds := &fileDataSource{path: absolutePath, dir: dir, enableWatch: watch}
	if watch {
		ds.changed = make(chan struct{}, 1)
		xgo.Go(ds.watch)
	}
	return ds
}

// ReadConfig ...
func (fp *fileDataSource) ReadConfig() (content []byte, err error) {
	f, err := os.Open(fp.path)
	//关闭
	defer f.Close()
	//定义返回变量
	if err != nil {
		return nil, err
	}
	//获取文件名称
	fileName := f.Name()
	//是否是properties 暂时只支持properties
	if strings.HasSuffix(fileName, ".properties") {
		data, readErr := conf.ReadConfigKeyValue(f)
		if nil != readErr {
			return nil, readErr
		}
		byteData, unErr := json.Marshal(data)
		if nil != unErr {
			return nil, unErr
		}
		return byteData, nil
	}
	return ioutil.ReadFile(fp.path)
}

// Close ...
func (fp *fileDataSource) Close() error {
	close(fp.changed)
	return nil
}

// IsConfigChanged ...
func (fp *fileDataSource) IsConfigChanged() <-chan struct{} {
	return fp.changed
}

// Watch file and automate update.
func (fp *fileDataSource) watch() {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		logrus.Fatal("new file watcher", logger.FieldMod("file datasource"), logger.Any("err", err.Error()))
	}

	defer w.Close()
	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-w.Events:
				logrus.Debug("read watch event",
					logger.FieldMod("file datasource"),
					logger.String("event", filepath.Clean(event.Name)),
					logger.String("path", filepath.Clean(fp.path)),
				)
				// we only care about the config file with the following cases:
				// 1 - if the config file was modified or created
				// 2 - if the real path to the config file changed
				const writeOrCreateMask = fsnotify.Write | fsnotify.Create
				if event.Op&writeOrCreateMask != 0 && filepath.Clean(event.Name) == filepath.Clean(fp.path) {
					logrus.Println("modified file: ", event.Name)
					select {
					case fp.changed <- struct{}{}:
					default:
					}
				}
			case err := <-w.Errors:
				// log.Println("error: ", err)
				logrus.Error("read watch error", logger.FieldMod("file datasource"), logger.Any("err", err.Error()))
			}
		}
	}()

	err = w.Add(fp.dir)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
