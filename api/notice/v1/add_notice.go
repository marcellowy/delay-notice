package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

type HttpNotice struct {
	Url     string            `json:"url"`
	Timeout int32             `json:"timeout"` // 回调时请求超时时间,单位: 秒
	Header  map[string]string `json:"header"`  // 回调时头的参数,没有就不填
}

type AddNoticeReq struct {
	g.Meta        `path:"/notice/add" tags:"" method:"post" summary:""`
	Type          string      `json:"type"`          // 被通知时的调用,默认: http|https
	Http          *HttpNotice `json:"http"`          // http通知方法
	Delay         int32       `json:"delay"`         // 延迟多久,单位:秒
	RetryTimes    int32       `json:"retry_times"`   // 通知失败时重试几次
	RetryInterval int32       `json:"retryInterval"` // 通知失败时重试间隔,单位:秒
	Data          string      `json:"data"`          // 通知时原样返回的数据
}

type AddNoticeRes struct {
	g.Meta   `mime:"text/html" example:"string"`
	NoticeId string `json:"notice_id"`
}
