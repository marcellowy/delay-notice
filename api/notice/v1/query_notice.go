// Package v1
// Copyright 2016-2024 chad.wang<chad.wang@icloudsky.com>. All rights reserved.
package v1

import "github.com/gogf/gf/v2/frame/g"

// NoticeStatus 通知状态
type NoticeStatus int32

const (
	NoticeStatusUnknown NoticeStatus = iota
	NoticeStatusSuccess
	NoticeStatusFailed
	NoticeStatusDoing
)

type QueryNoticeReq struct {
	g.Meta   `path:"/notice/query" tags:"" method:"get" summary:""`
	NoticeId string `json:"notice_id"`
}

type QueryNoticeRes struct {
	g.Meta  `mime:"text/html" example:"string"`
	Status  NoticeStatus `json:"status"`  // 消息状态; 0:等待被通知 1:通知成功 2:通知失败 3:正在执行通知
	Message string       `json:"message"` // 状态描述
}
