package brian

import (
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/brian-god/brian-go/pkg/client/xnacos_client"
	"github.com/brian-god/brian-go/pkg/conf"
	file_datasource "github.com/brian-god/brian-go/pkg/datasource/file"
	http_datasource "github.com/brian-god/brian-go/pkg/datasource/http"
	"github.com/brian-god/brian-go/pkg/discover"
	"github.com/brian-god/brian-go/pkg/discover/nacos_discover"
	"github.com/brian-god/brian-go/pkg/group"
	"github.com/brian-god/brian-go/pkg/logger"
	"github.com/brian-god/brian-go/pkg/registry"
	"github.com/brian-god/brian-go/pkg/registry/xnacos_registry"
	"github.com/brian-god/brian-go/pkg/server"
	"github.com/brian-god/brian-go/pkg/server/xgrpc"
	"github.com/brian-god/brian-go/pkg/server/xhttp"
	"github.com/brian-god/brian-go/pkg/utils/xgo"
	"github.com/brian-god/brian-go/pkg/worker"
	"github.com/brian-god/brian-go/pkg/xcast"
	"github.com/brian-god/brian-go/pkg/xcodec"
	"github.com/brian-god/brian-go/pkg/xfile"
	"github.com/brian-god/brian-go/pkg/xflag"
	"github.com/labstack/gommon/color"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"strings"
	"sync"
)

// Application is the framework's instance, it contains the servers, workers, client and configuration settings.
// Create an instance of Application, by using &Application{}
type Application struct {
	servers     []server.Server
	workers     []worker.Worker
	logger      *logrus.Logger
	stopOnce    sync.Once
	initOnce    sync.Once
	startupOnce sync.Once

	//registerer registry.Registry

	signalHooker func(*Application)
	defers       []func() error

	governor *http.Server
	colorer  *color.Color

	httpServer           *xhttp.Server
	rpcServer            *xgrpc.Server
	registry             registry.Registry
	discover             discover.Discover
	registryConfig       *registry.RegistryConfig
	Name                 string `properties:"brian.application.name"`                  //应用名称
	LogLevel             string `properties:"brian.application.log.level"`             // 日志级别
	EnableRpcServer      bool   `properties:"brian.application.enable.RpcServer"`      //是否开启rpc服务
	EnableRegistryCenter bool   `properties:"brian.application.enable.RegistryCenter"` //是否启用注册中心
}

// 初始化应用
func (app *Application) initialize() {
	app.initOnce.Do(func() {
		app.servers = make([]server.Server, 0)
		app.workers = make([]worker.Worker, 0)
		app.signalHooker = hookSignals
		app.defers = []func() error{}
	})
}

// 获取默认的应用
func DefaultApplication() *Application {
	//开始使用默认的
	app := &Application{colorer: color.New(), logger: logrus.New(), Name: "brian", LogLevel: "info", EnableRpcServer: false, EnableRegistryCenter: false}
	//打印logo
	app.printBanner()
	//调用应用初始化的方法
	app.initialize()
	//默认启动方式不进行配置加载
	//app.loadConfig()
	//创建http服务
	app.defaultServerHTTP()
	//判断是否需要开启rpc服务
	if app.EnableRpcServer {
		//创建rpc服务
		app.defaultServeGRPC()
	}
	return app
}

// 创建一个读取配置文件的application
func RewConfigApplication() *Application {
	//开始使用默认的
	app := &Application{colorer: color.New(), logger: logrus.New(), Name: "brian", LogLevel: "info", EnableRpcServer: false, EnableRegistryCenter: false}
	//打印logo
	app.printBanner()
	//调用应用初始化的方法
	app.initialize()
	//配置加载
	app.loadConfig()
	//读取配置文件中对应用的配置
	if err := conf.Unmarshal(app); err != nil {
		logrus.Panic("read app config  error", logger.FieldMod(xcodec.ModApp), logger.FieldErrKind(xcodec.ReadAppConfigErr), logger.FieldErr(err))
	}
	//创建读取配置http服务
	app.serverHTTP()
	//判断是否需要开启rpc服务
	if app.EnableRpcServer {
		//创建读取配置rpc服务
		app.serveGRPC()
	}
	//设置应用的日志级别
	if level, err := logrus.ParseLevel(app.LogLevel); nil == err {
		app.logger.Level = level
	}
	//判断应用是否开启注册中心
	if app.EnableRegistryCenter {
		//初始化配置中心
		app.registryCenter()
		//进行服务的注册
		//app.registryServer()
	}
	return app
}

//默认配置启动rpc服务
func (app *Application) defaultServeGRPC() error {
	//获取一个grpc服务
	rpcServer := xgrpc.DefaultConfig().Build()
	app.rpcServer = rpcServer
	return app.Serve(rpcServer)
}

//rpc服务
func (app *Application) serveGRPC() error {
	//获取一个grpc服务
	rpcServer := xgrpc.StdConfig().Build()
	app.rpcServer = rpcServer
	return app.Serve(rpcServer)
}

// RegisterRpcServer 注册rpc服务
func (app *Application) RegisterRpcServer(in interface{}, srv interface{}) {
	app.rpcServer.Register(in, srv)
}

//RegisterController 注册controller
func (app *Application) RegisterController(con xhttp.Controller) {
	app.httpServer.UseController(con)
}

//使用默认配置启动http服务
func (app *Application) defaultServerHTTP() error {
	httpServer := xhttp.DefaultConfig().Build()
	app.httpServer = httpServer
	return app.Serve(httpServer)
}

//http服务
func (app *Application) serverHTTP() error {
	httpServer := xhttp.StdConfig("http").Build()
	app.httpServer = httpServer
	return app.Serve(httpServer)
}

// 启动应用内部方法
func (app *Application) startup() (err error) {
	//执行注入的函数
	app.startupOnce.Do(func() {
		err = xgo.SerialUntilError(
			//放入执行的函数
			app.initLogger,
		)()
	})
	return
}

//创建注册中心
func (app *Application) registryCenter() {
	registryConfig, err := registry.RewConfig()
	//是否的能够获取到配置文件
	if nil != err {
		app.logger.Panic(logger.FieldMod(xcodec.ModRegistry), logger.FieldErrKind(xcodec.ReadRegistryConfigErr), logger.FieldErr(err))
	}
	//将注册中心的配置信息放入应用中
	app.registryConfig = registryConfig
	//获取配置中心的类型
	if xcodec.Nacos == registryConfig.Type {
		//nacos
		nacosConfig := xnacos_client.NewNacosClientConfig(registryConfig)
		nacosServerConfig := xnacos_registry.NacosServerConfigs(registryConfig)
		//创建一个nacos的client
		nacosClient, err1 := xnacos_client.NewNacosClient(nacosConfig, nacosServerConfig)
		if nil != err1 {
			app.logger.Panic("create nacos client error ", logger.FieldMod(xcodec.ModConfig), logger.FieldErr(err))
		}
		//获取注册中心
		app.registry = xnacos_registry.CreateNacosRegister(nacosClient)
		app.discover = nacos_discover.CreateNacoseDiscover(nacosClient)
	}
}

/*//注册服务
func (app *Application) registryServer()  {
	//注册http服务
	httpSeverConfig := app.httpServer.Config
	//获取服务名称
	httpServerName := httpSeverConfig.Name
	if httpServerName == ""{
		httpServerName = app.Name
	}
	registryConfig := app.registryConfig
	httpParam := &server.ServiceInfo{
		Name:httpServerName,
		Scheme:httpServerName,
		IP :httpSeverConfig.Host,
		Port:httpSeverConfig.Port,
		Weight:httpSeverConfig.Weight,
		Enable:true,
		Healthy:true,
		Ephemeral:true,
		GroupName:registryConfig.GroupName,
		ClusterName:registryConfig.ClusterName,
	}
	//注册服务
	if err:=app.registry.RegisterService(context.Background(),httpParam);err != nil {
		app.logger.Panic("registry http server error ", logger.FieldMod(xcodec.ModRegistry), logger.FieldErr(err))
	}
	//启用了rpc服务才进行rpc服务的注册
	if app.EnableRpcServer {
		//注册rpc服务
		rpcServer := app.rpcServer
		//获取服务名称
		rpcServerName := rpcServer.Name
		if httpServerName == ""{
			rpcServerName = app.Name+"-rpc"
		}
		rpcParam := &server.ServiceInfo{
			Name:rpcServerName,
			Scheme:rpcServerName,
			IP :rpcServer.Host,
			Port:rpcServer.Port,
			Weight:rpcServer.Weight,
			Enable:true,
			Healthy:true,
			Ephemeral:true,
			GroupName:registryConfig.GroupName,
			ClusterName:registryConfig.ClusterName,
		}
		//注册服务
		if err:=app.registry.RegisterService(context.Background(),rpcParam);err != nil {
			app.logger.Panic("registry http server error ", logger.FieldMod(xcodec.ModRegistry), logger.FieldErr(err))
		}
	}
}

//注销册服务
func (app *Application) deregisterService()  {
	//注http服务
	httpSeverConfig := app.httpServer.Config
	//获取服务名称
	httpServerName := httpSeverConfig.Name
	if httpServerName == ""{
		httpServerName = app.Name
	}
	registryConfig := app.registryConfig
	httpParam := &server.ServiceInfo{
		Name:httpServerName,
		Scheme:httpServerName,
		IP :httpSeverConfig.Host,
		Port:httpSeverConfig.Port,
		Weight:httpSeverConfig.Weight,
		Enable:true,
		Healthy:true,
		Ephemeral:true,
		GroupName:registryConfig.GroupName,
		ClusterName:registryConfig.ClusterName,
	}
	//注销服务
	if err:=app.registry.DeregisterService(context.Background(),httpParam);err != nil {
		app.logger.Panic("deregister http server error ", logger.FieldMod(xcodec.ModRegistry), logger.FieldErr(err))
	}
	//启用了rpc服务才进行rpc服务的注销
	if app.EnableRpcServer {
		//注销rpc服务
		rpcServer := app.rpcServer
		//获取服务名称
		rpcServerName := rpcServer.Name
		if httpServerName == ""{
			rpcServerName = app.Name+"-rpc"
		}
		rpcParam := &server.ServiceInfo{
			Name:rpcServerName,
			Scheme:rpcServerName,
			IP :rpcServer.Host,
			Port:rpcServer.Port,
			Weight:rpcServer.Weight,
			Enable:true,
			Healthy:true,
			Ephemeral:true,
			GroupName:registryConfig.GroupName,
			ClusterName:registryConfig.ClusterName,
		}
		//注销服务
		if err:=app.registry.DeregisterService(context.Background(),rpcParam);err != nil {
			app.logger.Panic("deregister http server error ", logger.FieldMod(xcodec.ModRegistry), logger.FieldErr(err))
		}
	}
}*/
func (app *Application) initLogger() error {
	logrus.SetOutput(os.Stdout)
	//日志级别
	if v := conf.Get(xcodec.ApplicationLoglevel); v != nil {
		if v, err := xcast.ToStringE(v); nil == err {
			if level, err := logrus.ParseLevel(v); nil == err {
				logrus.SetLevel(level)
			}
		}
	}
	logrus.Debug("debug 日志")
	return nil
}

//提供外部启动应用执行
func (app *Application) Startup(fns ...func() error) error {
	if err := app.startup(); err != nil {
		return err
	}
	return xgo.SerialUntilError(fns...)()
}

// GracefulStop 完成必要的清理后停止应用程序
func (app *Application) GracefulStop(ctx context.Context) (err error) {
	app.beforeStop()
	app.stopOnce.Do(func() {
		//清理注册中心
		err = app.registry.Close()
		if err != nil {
			app.logger.Errorf("graceful stop register close err", logger.FieldMod(xcodec.ModApp), logger.FieldErr(err))
		}
		/*err = app.governor.Close()
		if err != nil {
			app.logger.Error("graceful stop governor close err", xlog.FieldMod(ecode.ModApp), xlog.FieldErr(err))
		}*/
		var eg errgroup.Group
		//停止http服务
		/*if app.httpServer != nil {
			eg.Go(func() error {
				return app.httpServer.GracefulStop(ctx)
			})
		}
		//停止rpc服务
		if app.rpcServer != nil {
			eg.Go(func() error {
				return app.rpcServer.GracefulStop(ctx)
			})
		}*/
		for _, s := range app.servers {
			s := s
			eg.Go(func() error {
				return s.GracefulStop(ctx)
			})
		}
		err = eg.Wait()
	})
	return err
}

// Stop 完成必要的清理后立即停止程序
func (app *Application) Stop() (err error) {
	app.beforeStop()
	app.stopOnce.Do(func() {
		//清理注册中心
		/*err = app.registerer.Close()
		if err != nil {
			app.logger.Error("stop register close err", xlog.FieldMod(ecode.ModApp), xlog.FieldErr(err))
		}
		err = app.governor.Close()
		if err != nil {
			app.logger.Error("stop governor close err", xlog.FieldMod(ecode.ModApp), xlog.FieldErr(err))
		}*/
		var eg errgroup.Group
		//停止http服务
		/*if app.httpServer != nil {
			eg.Go(app.httpServer.Stop)
		}
		//停止rpc服务
		if app.rpcServer != nil {
			eg.Go(app.rpcServer.Stop)
		}*/
		for _, s := range app.servers {
			s := s
			eg.Go(s.Stop)
		}
		for _, w := range app.workers {
			w := w
			eg.Go(w.Stop)
		}
		err = eg.Wait()
	})
	return
}

// Run run application
func (app *Application) Run() error {
	defer app.clean()
	if app.signalHooker == nil {
		app.signalHooker = hookSignals
	}
	/*if app.governor == nil {
		app.governor = &http.Server{
			Handler: govern.DefaultServeMux,
			Addr:    "127.0.0.1:9990", // 默认治理端口
		}
	}*/
	//注册
	if app.registry == nil {
		app.registry = registry.Nop{}
	}

	app.signalHooker(app)

	// start govern
	var eg errgroup.Group
	//eg.Go(app.startGovernor)
	eg.Go(app.startServers)
	eg.Go(app.startWorkers)
	return eg.Wait()
}

//开启工作线程
func (app *Application) startWorkers() error {
	var eg group.Group
	// start multi workers
	for _, w := range app.workers {
		w := w
		eg.Go(func() error {
			return w.Run()
		})
	}
	return eg.Wait()
}

// 启动服务
func (app *Application) startServers() error {
	registryConfig := app.registryConfig
	var eg errgroup.Group
	//启动http服务
	/*if app.httpServer != nil {
		eg.Go(func() (err error) {
			return app.httpServer.Serve()
		})
	}
	//启动rpc服务
	if app.rpcServer != nil {
		eg.Go(func() (err error) {
			return app.rpcServer.Serve()
		})
	}*/
	xgo.ParallelWithErrorChan()
	// start multi servers
	for _, s := range app.servers {
		s := s
		eg.Go(func() (err error) {
			serverInfo := s.Info(registryConfig.GroupName, registryConfig.ClusterName)
			//注册服务
			_ = app.registry.RegisterService(context.TODO(), serverInfo)
			//注销服务
			//defer app.registry.DeregisterService(context.TODO(), serverInfo)
			//defer app.registry.DeregisterService(context.TODO(), serverInfo)
			app.logger.Info("start servers", logger.FieldMod(xcodec.ModApp), logger.FieldAddr(serverInfo.Label()), logger.Any("scheme", serverInfo.Scheme))
			//defer app.logger.Info("exit server", logger.FieldMod(xcodec.ModApp), logger.FieldErr(err), logger.FieldAddr(serverInfo.Label()))
			return s.Serve()
		})
	}
	return eg.Wait()
}

func (app *Application) clean() {
	for i := len(app.defers) - 1; i >= 0; i-- {
		fn := app.defers[i]
		if err := fn(); err != nil {
			//xlog.Error("clean.defer", xlog.String("func", xstring.FunctionName(fn)))
		}
	}
	//_ = xlog.DefaultLogger.Flush()
	//_ = xlog.JupiterLogger.Flush()
}
func (app *Application) beforeStop() {
	if app.EnableRegistryCenter {
		app.logger.Info("停止服务并注销注册中心的服务")
		//注销服务
		app.deregisterService()
	}
	// 应用停止之前的处理
	//app.logger.Info("leaving jupiter, bye....", xlog.FieldMod(ecode.ModApp))
}

//deregisterService 进行服务注册
func (app *Application) deregisterService() {
	registryConfig := app.registryConfig
	for _, s := range app.servers {
		//获取服务信息
		serverInfo := s.Info(registryConfig.GroupName, registryConfig.ClusterName)
		//注销服务
		err := app.registry.DeregisterService(context.TODO(), serverInfo)
		app.logger.Info("exit server", logger.FieldMod(xcodec.ModApp), logger.FieldErr(err), logger.FieldAddr(serverInfo.Label()))
	}
}

//注册服务
func (app *Application) Serve(s server.Server) error {
	app.servers = append(app.servers, s)
	return nil
}

//加载配置
func (app *Application) loadConfig() error {
	var (
		watchConfig = xflag.Bool("watch")
		configAddr  = xflag.String("config")
	)

	if configAddr == "" {
		app.logger.Warn("no config ... read default config")
		//为空则读取默认文件
		//优先级
		//botostrop.yml
		//application.yml
		//application.properties
		dir, _ := os.Getwd()
		ok, _ := xfile.PathExists(fmt.Sprintf("%s/resources/botostrop.yml", dir))
		if !ok {
			ok, _ = xfile.PathExists(fmt.Sprintf("%s/resources/application.yml", dir))
			if !ok {
				ok, _ = xfile.PathExists(fmt.Sprintf("%s/resources/application.properties", dir))
				if !ok {
					return nil
				} else {
					conf.SetConfigType("properties")
					configAddr = fmt.Sprintf("%s/resources/application.properties", dir)
				}
			} else {
				conf.SetConfigType("yml")
				configAddr = fmt.Sprintf("%s/resources/application.yml", dir)
			}
		} else {
			conf.SetConfigType("yml")
			configAddr = fmt.Sprintf("%s/botostrop.yml", dir)
		}
	}
	switch {
	case strings.HasPrefix(configAddr, "http://"),
		strings.HasPrefix(configAddr, "https://"):
		provider := http_datasource.NewDataSource(configAddr, watchConfig)
		if err := conf.LoadFromDataSource(provider, toml.Unmarshal); err != nil {
			app.logger.Panic("load remote config ", logger.FieldMod(xcodec.ModConfig), logger.FieldErrKind(xcodec.ErrKindUnmarshalConfigErr), logger.FieldErr(err))
		}
		app.logger.Info("load remote config ", logger.FieldMod(xcodec.ModConfig), logger.FieldAddr(configAddr))
	default:
		provider := file_datasource.NewDataSource(configAddr, watchConfig)

		if err := conf.LoadFromDataSource(provider, conf.UnmarshallerKeyAndValue); err != nil {
			app.logger.Panic("load local file ", logger.FieldMod(xcodec.ModConfig), logger.FieldErrKind(xcodec.ErrKindUnmarshalConfigErr), logger.FieldErr(err))
		}
		app.logger.Info("load local file ", logger.FieldMod(xcodec.ModConfig), logger.FieldAddr(configAddr))
	}
	return nil
}
func (app *Application) printBanner() error {
	const banner = `
	          _____                    _____                    _____                   _______
	         /\    \                  /\    \                  /\    \                 /::\    \
	        /::\____\                /::\____\                /::\    \               /::::\    \
	       /:::/    /               /:::/    /               /::::\    \             /::::::\    \
	      /:::/    /               /:::/    /               /::::::\    \           /::::::::\    \
	     /:::/    /               /:::/    /               /:::/\:::\    \         /:::/~~\:::\    \
	    /:::/____/               /:::/    /               /:::/  \:::\    \       /:::/    \:::\    \
	   /::::\    \              /:::/    /               /:::/    \:::\    \     /:::/    / \:::\    \
	  /::::::\    \   _____    /:::/    /      _____    /:::/    / \:::\    \   /:::/____/   \:::\____\
	 /:::/\:::\    \ /\    \  /:::/____/      /\    \  /:::/    /   \:::\ ___\ |:::|    |     |:::|    |
	/:::/  \:::\    /::\____\|:::|    /      /::\____\/:::/____/  ___\:::|    ||:::|____|     |:::|    |
	\::/    \:::\  /:::/    /|:::|____\     /:::/    /\:::\    \ /\  /:::|____| \:::\    \   /:::/    /
	 \/____/ \:::\/:::/    /  \:::\    \   /:::/    /  \:::\    /::\ \::/    /   \:::\    \ /:::/    /
	          \::::::/    /    \:::\    \ /:::/    /    \:::\   \:::\ \/____/     \:::\    /:::/    /
	           \::::/    /      \:::\    /:::/    /      \:::\   \:::\____\        \:::\__/:::/    /
	           /:::/    /        \:::\__/:::/    /        \:::\  /:::/    /         \::::::::/    /
	          /:::/    /          \::::::::/    /          \:::\/:::/    /           \::::::/    /
	         /:::/    /            \::::::/    /            \::::::/    /             \::::/    /
	        /:::/    /              \::::/    /              \::::/    /               \::/____/
	        \::/    /                \::/____/                \::/____/                 ~~
	         \/____/                  ~~

	 Welcome to hugo, starting application ...
	`
	/*const banner = `
	  o
	 <|>
	 / >
	 \o__ __o     o       o     o__ __o/    o__ __o
	  |     v\   <|>     <|>   /v     |    /v     v\
	 / \     <\  < >     < >  />     / \  />       <\
	 \o/     o/   |       |   \      \o/  \         /
	  |     <|    o       o    o      |    o       o
	 / \    / \   <\__ __/>    <\__  < >   <\__ __/>
									  |
							  o__     o
							  <\__ __/>

	Welcome to hugo, starting application ...
	`*/
	if app.colorer == nil {
		app.colorer = color.New()
	}
	app.colorer.Printf("%s\n", app.colorer.Blue(banner))
	return nil
}
