// Package notice
// Copyright 2016-2024 chad.wang<chad.wang@icloudsky.com>. All rights reserved.
package notice

import (
	"context"
	"delay-notice/internal/dao"
	"delay-notice/internal/model/entity"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/marcellowy/go-common/gogf/vlog"
	"github.com/marcellowy/go-common/tools"
	"sync"
	"time"
)

var (
	DataNoticeIdChan = make(chan *DataNoticeIdChanData, 0xfffff)
)

type DataNoticeIdChanData struct {
	Wg       *sync.WaitGroup
	NoticeId string
}

type Timer struct {
	PoolSize       int
	ScanDbInterval time.Duration
}

// 扫描要处理的数据
func (*Timer) scan(ctx context.Context) {

	var (
		rows []*entity.NoticeTimer
		err  error
	)

	vlog.Debug(ctx, "begin scan database")
	err = dao.DB.Where("prepare_notice_time < ? and status = ?", time.Now().Format(tools.TimeFormatString), 0).
		Limit(10000).
		Find(&rows).Error

	if err != nil {
		vlog.Error(ctx, err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(rows))

	vlog.Debug(ctx, "read data size: ", len(rows))
	for _, row := range rows {
		vlog.Debug(ctx, "put notice_id to queue: ", row.NoticeId)
		DataNoticeIdChan <- &DataNoticeIdChanData{Wg: &wg, NoticeId: row.NoticeId}
	}

	wg.Wait()
}

// startPool 启动处理协程
func (t *Timer) startPool(ctx context.Context, size int) {
	for index := 0; index < size; index++ {
		vlog.Infof(ctx, "start goroutine %d", index)
		go func(index int) {
			for {
				select {
				case dd, ok := <-DataNoticeIdChan:
					if !ok {
						break
					}
					cxx := gctx.New()
					vlog.Infof(cxx, "goroutine %d start work process notice_id: %s", index, dd.NoticeId)
					var beginTime = time.Now()

					Cb(cxx, dd.NoticeId)
					dd.Wg.Done()

					// 打印处理一个文件的耗时
					vlog.Infof(cxx, "goroutine %d process success; process notice_id: %s; cost: %s",
						index, dd.NoticeId, time.Now().Sub(beginTime).String())
				}
			}
		}(index)
	}
}

func (t *Timer) Start(ctx context.Context) {
	// 启动处理协程
	t.startPool(ctx, t.PoolSize)

	go func() {
		for {
			select {
			case <-time.After(t.ScanDbInterval):
				cxx := gctx.New()
				t.scan(cxx)
			}
		}
	}()
}
