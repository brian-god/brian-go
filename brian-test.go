package brian

import (
	"cloud.google.com/go/trace/testdata/helloworld"
	"context"
	"fmt"
	"github.com/brian-god/brian-go/pkg/server/xhttp"
	"github.com/labstack/echo/v4"
	"net/http"
)

var ser = Application{}

//构建一个controller

type TestController struct {
}

func (test *TestController) Register(server *xhttp.Server) {
	server.GET("/index", test.index)
}

//写入一个测试的方法
func (test *TestController) index(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, "hello hugo")
}

func main() {

	/*r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	go r.Run() // listen and serve on 0.0.0.0:8080
	server := grpc.NewServer()

	helloworld.RegisterGreeterServer(server, new(Greeter))

	lis, err := net.Listen("tcp", ":8090")
	if err != nil {
		panic(err.Error())
	}
	server.Serve(lis)*/

	if err := ser.Startup(serverHTTP, serveGRPC); err != nil {
		fmt.Println("启动有误")
	}
	ser.Run()
}
func Hello() error {
	fmt.Printf("你好")
	return nil
}

//rpc服务
func serveGRPC() error {
	//获取grpc服务
	//grpcServer  := xgrpc.DefaultConfig().Build()
	return nil
}
func serverHTTP() error {
	httpServer := xhttp.StdConfig("http").Build()
	//使用
	httpServer.UseController(&TestController{})
	//启动服务
	httpServer.Serve()
	return nil
}

type Greeter struct {
	helloworld.GreeterServer
}

//grpc
func (g Greeter) SayHello(context context.Context, request *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return &helloworld.HelloReply{
		Message: "Hello Jupiter",
	}, nil
}
