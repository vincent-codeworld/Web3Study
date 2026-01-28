package main

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func main() {
	// 1. 初始化默认服务器
	// Default() 会自带恢复中间件（Recovery），默认监听 8888 端口
	h := server.Default()
	// 2. 定义一个简单的 GET 路由
	// 访问: http://localhost:8888/ping
	h.Group("")
	h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, utils.H{
			"message": "pong",
		})
	})

	// 3. 定义一个带路径参数的路由
	// 访问: http://localhost:8888/user/zhangsan
	h.GET("/user/:name", func(c context.Context, ctx *app.RequestContext) {
		// 获取 URL 中的参数
		name := ctx.Param("name")

		ctx.JSON(consts.StatusOK, utils.H{
			"user":      name,
			"status":    "active",
			"framework": "hertz",
		})
	})

	// 4. 启动服务
	// Spin() 会阻塞主 goroutine 并开始监听
	h.Spin()
}
