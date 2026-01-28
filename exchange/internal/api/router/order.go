package router

import "github.com/cloudwego/hertz/pkg/route"

func RegisterOrderRouter(group *route.RouterGroup) {
	orderGroup := group.Group("/order")
	orderGroup.POST("/create")
}
