// Package notice
// Copyright 2016-2024 chad.wang<chad.wang@icloudsky.com>. All rights reserved.
package notice

import (
	"context"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/marcellowy/go-common/gogf/vlog"
	"time"
)

// poolStart 多线程处理文件
func poolStart(ctx context.Context, poolSize int) {
	// 使用一次性创建N个线程，然后N个线程读取同一个chan的方式，避免多次创建goroutine的开销
	for i := 0; i < poolSize; i++ {
		go consumerPool(ctx, i)
	}
}

// consumerPool 启动线程进行数据的处理
func consumerPool(ctx context.Context, index int) {
	vlog.Infof(ctx, "start goroutine: %d", index)
	for {
		select {
		case noticeId, ok := <-PoolData:
			if !ok {
				break // channel closed maby
			}
			cxx := gctx.New() // 创建一个新的ctx，避免日志混乱
			vlog.Infof(cxx, "goroutine %d start work process notice_id: %s", index, noticeId)
			var beginTime = time.Now()

			Callback(cxx, noticeId)

			// 打印处理一个文件的耗时
			vlog.Infof(cxx, "goroutine %d process success; process notice_id: %s; cost: %s",
				index, noticeId, time.Now().Sub(beginTime).String())
		}
	}
}
