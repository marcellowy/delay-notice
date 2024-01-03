// Package auto_create_table
// Copyright 2016-2024 chad.wang<chad.wang@icloudsky.com>. All rights reserved.
package auto_create_table

import (
	"context"
	"delay-notice/internal/dao"
	"delay-notice/internal/model/entity"
	"fmt"
	"github.com/gogf/gf/v2/os/gcron"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/marcellowy/go-common/gogf/vlog"
	"gorm.io/gorm"
	"time"
)

func createTable(ctx context.Context, db *gorm.DB, day int) {
	var err error
	name := fmt.Sprintf("t_notice_%s", time.Now().AddDate(0, 0, day).Format("20060102"))
	if err = db.Table(name).AutoMigrate(&entity.Notice{}); err != nil {
		vlog.Error(ctx, err)
		return
	}
	vlog.Infof(ctx, "create %s success", name)
}

func start(ctx context.Context, db *gorm.DB, day int) {
	if day < 0 {
		for i := 0; i > day; i-- {
			createTable(ctx, db, i)
		}
	}

	for i := 0; i < day; i++ {
		createTable(ctx, db, i)
	}
}

// CreateTableNotice 每天凌晨创建表
func CreateTableNotice() {
	var (
		err error
		ctx = gctx.New()
	)

	// 添加单例定时任务
	// 每天凌晨两点
	var entry *gcron.Entry
	entry, err = gcron.AddSingleton(ctx, "0 0 2 * * *", func(ctx context.Context) {
		startTime := time.Now()
		vlog.Debugf(ctx, "create table at %s", startTime.Format(time.RFC3339))

		start(ctx, dao.DB, 3)

		endTime := time.Now()
		vlog.Debugf(ctx, "create table end at %s cost: %.2fs", endTime.Format(time.RFC3339), endTime.Sub(startTime).Seconds())
	})

	if err != nil {
		vlog.Error(ctx, err)
		return
	}

	vlog.Debug(ctx, "create table at", time.Now().Format(time.RFC3339))
	entry.Start()

	start(ctx, dao.DB, -3)
	start(ctx, dao.DB, 3)
}
