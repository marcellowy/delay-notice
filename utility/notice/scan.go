// Package notice
// Copyright 2016-2024 chad.wang<chad.wang@icloudsky.com>. All rights reserved.
package notice

import (
	"context"
	"delay-notice/internal/dao"
	"delay-notice/internal/model/entity"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/marcellowy/go-common/gogf/vlog"
	"github.com/marcellowy/go-common/tools"
	"time"
)

var (
	// PoolData 池数据
	PoolData = make(chan string, 0xffff)

	// TimerDataFailed 处理失败的数据
	TimerDataFailed = make(chan *entity.Notice, 0xffff)

	// TimerDataSuccess 处理成功的数据
	TimerDataSuccess = make(chan string, 0xffff)
)

// QueryAndWriteChan 查询数据并放入chan
func QueryAndWriteChan(ctx context.Context) {
	var (
		rows []*entity.NoticeTimer
		err  error
	)
	err = dao.DB.Where("prepare_notice_time < ? and status = 0", time.Now().Format(tools.TimeFormatString)).
		Find(&rows).Error

	if err != nil {
		vlog.Error(ctx, err)
		return
	}

	for _, row := range rows {
		PoolData <- row.NoticeId
		// 更新数据表
		err = dao.DB.Model(&entity.NoticeTimer{}).
			Where("id = ?", row.ID).
			Updates(map[string]interface{}{
				"status": 1,
			}).Error
		if err != nil {
			vlog.Error(ctx, err)
			continue
		}
	}
}

// ProcessFailedData 处理需要重试的数据
func ProcessFailedData(ctx context.Context, data *entity.Notice) {

	var nt = time.Now().Add(time.Second * time.Duration(data.RetryInterval))

	if err := dao.DB.Model(&entity.NoticeTimer{}).Where("notice_id = ?", data.NoticeId).Updates(map[string]interface{}{
		"status":              0,
		"prepare_notice_time": nt,
	}).Error; err != nil {
		vlog.Error(ctx, err)
	}
}

// Start 扫描数据库
func Start(ctx context.Context) {

	go func() {

		var interval = g.Config().MustGet(ctx, "notice.scanDatabaseInterval").Int32()

		for {
			cxx := gctx.New()

			select {

			case <-time.After(time.Second * time.Duration(interval)):
				// 扫描数据进行处理
				QueryAndWriteChan(cxx)

			case noticeId, ok := <-TimerDataSuccess:
				// 已经确定成功的数据
				if !ok {
					continue
				}

				// 真实删除
				if err := dao.DB.Unscoped().Where("notice_id = ?", noticeId).Delete(&entity.NoticeTimer{}).Error; err != nil {
					vlog.Error(cxx, err)
				}
			case data, ok := <-TimerDataFailed:
				if !ok {
					continue
				}
				ProcessFailedData(cxx, data)
			}
		}
	}()

	var number = g.Config().MustGet(ctx, "notice.processDataGoroutine").Int()
	if number <= 0 {
		number = 10
	}

	// 处理协程
	poolStart(ctx, number)

	// 兜底协程
}
