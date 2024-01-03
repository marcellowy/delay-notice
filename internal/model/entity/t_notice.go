// Package entity
// Copyright 2016-2024 chad.wang<chad.wang@icloudsky.com>. All rights reserved.
package entity

import (
	"fmt"
	"github.com/marcellowy/go-common/tools"
	"gorm.io/gorm"
	"time"
)

type NoticeTIName struct {
	TableName string
	NoticeId  string
}

type Notice struct {
	gorm.Model
	NoticeId          string     `gorm:"type:varchar(24);not null;default:'';uniqueIndex"` // 通知Id
	Type              string     `gorm:"type:varchar(24);not null;default:''"`             // 通知协议类型
	TypeData          string     `gorm:"type:text"`                                        // 通知协议对应的需要的内容,这里有专用结构体来解析与api使用相同的结构体
	Delay             int32      `gorm:"type:int(10);not null;default:0"`                  // 延迟时间
	PrepareNoticeTime *time.Time // 预准备通知时间
	RealNoticeTime    *time.Time // 实际通知时间
	Status            int32      `gorm:"type:int(10);not null;default:0"` // 状态
	StatusDesc        string     `gorm:"type:text"`                       // 通知失败或者运行状态消息描述
	RetryTimes        int32      `gorm:"type:int(10);not null;default:0"` // 重试次数
	RetryInterval     int32      `gorm:"type:int(10);not null;default:0"` // 重试间隔时间
	RealRetryTimes    int32      `gorm:"type:int(10);not null;default:0"` // 实际重试次数
	Data              string     `gorm:"type:text"`                       // 通知数据
}

func (*Notice) TableName() string {
	return "t_notice"
}

// Generate 生成表名和id
func (n *Notice) Generate() *NoticeTIName {
	// notice_id
	var nt = NoticeTIName{}
	nowDay := time.Now().Format("20060102")
	nowSec := time.Now().Format("20060102150405")

	nt.TableName = fmt.Sprintf("t_notice_%s", nowDay)
	nt.NoticeId = nowSec + tools.RandomString(6, tools.RandomDigital)

	n.NoticeId = nt.NoticeId
	return &nt
}

// GetNoticeTIName 从已经存在的id推导出表名
func (n *Notice) GetNoticeTIName() *NoticeTIName {
	var nt = NoticeTIName{}
	nt.TableName = fmt.Sprintf("t_notice_%s", n.NoticeId[:8])
	nt.NoticeId = n.NoticeId

	return &nt
}
