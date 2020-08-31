module github.com/brian-god/brian-go

go 1.13

require (
	cloud.google.com/go v0.26.0
	github.com/BurntSushi/toml v0.3.1
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-resty/resty/v2 v2.3.0
	github.com/golang/protobuf v1.4.2
	github.com/labstack/echo/v4 v4.1.16 //一个微型的web框架
	github.com/labstack/gommon v0.3.0
	github.com/mitchellh/mapstructure v1.3.3 // indirect go结构体映射
	github.com/nacos-group/nacos-sdk-go v1.0.0
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron v1.2.0 //定时任务库
	github.com/sirupsen/logrus v1.6.0 //日志框架
	go.uber.org/zap v1.15.0
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	google.golang.org/grpc v1.30.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/yaml.v2 v2.3.0 // indirect
)
