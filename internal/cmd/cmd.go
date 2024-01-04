package cmd

import (
	"context"
	"delay-notice/internal/controller/notice"
	"delay-notice/internal/dao"
	"delay-notice/utility/auto_create_table"
	notice2 "delay-notice/utility/notice"
	"github.com/marcellowy/go-common/gogf/config"
	"github.com/marcellowy/go-common/gogf/middleware"
	"github.com/marcellowy/go-common/gogf/vlog"
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
			{
				i := config.Get("notice.processDataGoroutine").Int()
				if i <= 0 {
					i = 5
				}
				vlog.Info(ctx, "processDataGoroutine: ", i)

				sdInterval := time.Duration(config.Get("notice.scanDatabaseInterval").Int32()) * time.Second
				vlog.Info(ctx, "scanDatabaseInterval: ", sdInterval.String())

				timer := notice2.Timer{
					PoolSize:       i,
					ScanDbInterval: sdInterval,
				}
				timer.Start(ctx)
			}

			s.Run()
			return nil
		},
	}
)
