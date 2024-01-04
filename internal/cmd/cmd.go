package cmd

import (
	"context"
	"delay-notice/internal/controller/notice"
	"delay-notice/internal/dao"
	"delay-notice/utility/auto_create_table"
	notice2 "delay-notice/utility/notice"
	"github.com/marcellowy/go-common/gogf/middleware"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()
			s.Group("/api/v1", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Middleware(ghttp.MiddlewareCORS)
				group.Middleware(ghttp.MiddlewareJsonBody)
				group.Middleware(middleware.Print)
				group.Bind(
					notice.NewV1(),
				)
			})

			// 连接数据库
			dao.NewConnect(ctx)

			// 自动创建按天的任务
			auto_create_table.CreateTableNotice()

			// 创建其他表
			dao.CreateTable(ctx)

			// 启动处理通知协程
			timer := notice2.Timer{
				PoolSize:       10,
				ScanDbInterval: time.Second * 2,
			}
			timer.Start(ctx)

			s.Run()
			return nil
		},
	}
)
