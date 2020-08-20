package brian

import (
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/brian-god/brian-go/pkg/conf"
	file_datasource "github.com/brian-god/brian-go/pkg/datasource/file"
	http_datasource "github.com/brian-god/brian-go/pkg/datasource/http"
	"github.com/brian-god/brian-go/pkg/group"
	"github.com/brian-god/brian-go/pkg/logger"
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

	httpServer *xhttp.Server
	rpcServer  *xgrpc.Server
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
	app := &Application{colorer: color.New(), logger: logrus.New()}
	//打印logo
	app.printBanner()
	//加载配置
	app.loadConfig()
	//创建http服务
	app.serverHTTP()
	//创建rpc服务
	app.serveGRPC()
	return app
}

//rpc服务
func (app *Application) serveGRPC() error {
	//获取一个grpc服务
	rpcServer := xgrpc.StdConfig().Build()
	app.rpcServer = rpcServer
	return nil
}

// RegisterRpcServer 注册rpc服务
func (app *Application) RegisterRpcServer(in interface{}, srv interface{}) {
	app.rpcServer.Register(in, srv)
}

//RegisterController 注册controller
func (app *Application) RegisterController(con xhttp.Controller) {
	app.httpServer.UseController(con)
}

//http服务
func (app *Application) serverHTTP() error {
	httpServer := xhttp.StdConfig("http").Build()
	app.httpServer = httpServer
	return nil
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

func (app *Application) initLogger() error {
	logrus.SetOutput(os.Stdout)
	//日志级别
	if v := conf.Get("brian.application.log.level"); v != nil {
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
	//调用应用初始化的方法
	app.initialize()
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
		/*err = app.registerer.Close()
		if err != nil {
			app.logger.Error("graceful stop register close err", xlog.FieldMod(ecode.ModApp), xlog.FieldErr(err))
		}
		err = app.governor.Close()
		if err != nil {
			app.logger.Error("graceful stop governor close err", xlog.FieldMod(ecode.ModApp), xlog.FieldErr(err))
		}*/
		var eg errgroup.Group
		//停止http服务
		if app.httpServer != nil {
			eg.Go(func() error {
				return app.httpServer.GracefulStop(ctx)
			})
		}
		//停止rpc服务
		if app.rpcServer != nil {
			eg.Go(func() error {
				return app.rpcServer.GracefulStop(ctx)
			})
		}
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
		if app.httpServer != nil {
			eg.Go(app.httpServer.Stop)
		}
		//停止rpc服务
		if app.rpcServer != nil {
			eg.Go(app.rpcServer.Stop)
		}
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
	/*if app.registerer == nil {
		app.registerer = registry.Nop{}
	}*/

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
	var eg errgroup.Group
	//启动http服务
	if app.httpServer != nil {
		eg.Go(func() (err error) {
			return app.httpServer.Serve()
		})
	}
	//启动rpc服务
	if app.rpcServer != nil {
		eg.Go(func() (err error) {
			return app.rpcServer.Serve()
		})
	}
	//xgo.ParallelWithErrorChan()
	// start multi servers
	for _, s := range app.servers {
		s := s
		eg.Go(func() (err error) {
			//_ = app.registerer.RegisterService(context.TODO(), s.Info())
			//defer app.registerer.DeregisterService(context.TODO(), s.Info())
			//app.logger.Info("start servers", xlog.FieldMod(ecode.ModApp), xlog.FieldAddr(s.Info().Label()), xlog.Any("scheme", s.Info().Scheme))
			//defer app.logger.Info("exit server", xlog.FieldMod(ecode.ModApp), xlog.FieldErr(err), xlog.FieldAddr(s.Info().Label()))
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
	// 应用停止之前的处理
	//app.logger.Info("leaving jupiter, bye....", xlog.FieldMod(ecode.ModApp))
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
					configAddr = fmt.Sprintf("%s/resources/application.properties", dir)
				}
			} else {
				configAddr = fmt.Sprintf("%s/resources/application.yml", dir)
			}
		} else {
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
