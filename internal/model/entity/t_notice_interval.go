// Package entity
// Copyright 2016-2024 chad.wang<chad.wang@icloudsky.com>. All rights reserved.
package entity

import (
	"gorm.io/gorm"
	"time"
)

type NoticeTimer struct {
	gorm.Model
	NoticeId          string    `gorm:"type:varchar(24);not null;default:'';index"` // 通知Id
	PrepareNoticeTime time.Time `gorm:"index:idx_time_status"`                      // 预准备通知时间
	Status            int32     `gorm:"index:idx_time_status"`                      // 状态: 0:等待 1:已处理
}

func (*NoticeTimer) TableName() string {
	return "t_notice_timer"
}
