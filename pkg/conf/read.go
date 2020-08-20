package conf

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"strings"
)

/**
 *
 * Copyright (C) @2020 hugo network Co. Ltd
 * @description
 * @updateRemark
 * @author               hugo
 * @updateUser
 * @createDate           2020/8/19 5:28 下午
 * @updateDate           2020/8/19 5:28 下午
 * @version              1.0
**/

//反序列化
func UnmarshallerKeyAndValue(p []byte, v interface{}) error {
	return json.Unmarshal(p, v)
}

//读取key=value类型的配置文件(properties)
func ReadConfigKeyValue(file *os.File) (map[string]string, error) {
	config := make(map[string]string)
	r := bufio.NewReader(file)
	for {
		b, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		s := strings.TrimSpace(string(b))
		//判断是否以#开头的如果是则忽略掉
		if strings.HasPrefix(s, "#") {
			continue
		}
		index := strings.Index(s, "=")
		if index < 0 {
			continue
		}
		//判断是否包含行后面的注解
		i := strings.Index(s, "#")
		if i != -1 {
			s = s[:i]
		}
		key := strings.TrimSpace(s[:index])
		if len(key) == 0 {
			continue
		}
		value := strings.TrimSpace(s[index+1:])
		if len(value) == 0 {
			continue
		}
		config[key] = value
	}
	return config, nil
}
