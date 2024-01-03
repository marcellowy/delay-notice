// Package dao
// Copyright 2016-2024 chad.wang<chad.wang@icloudsky.com>. All rights reserved.
package dao

import (
	"context"
	"delay-notice/internal/model/entity"
	"github.com/marcellowy/go-common/gogf/db"
	"github.com/marcellowy/go-common/gogf/vlog"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
)

// NewConnect 连接数据库
func NewConnect(ctx context.Context) {
	DB = db.NewConnect(ctx, "database")
}

func CreateTable(ctx context.Context) {
	if err := DB.AutoMigrate(&entity.NoticeTimer{}); err != nil {
		vlog.Error(ctx, err)
	}
}
