// Package notice
// Copyright 2016-2024 chad.wang<chad.wang@icloudsky.com>. All rights reserved.
package notice

import (
	"context"
	v1 "delay-notice/api/notice/v1"
	"delay-notice/internal/dao"
	"delay-notice/internal/model/entity"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/marcellowy/go-common/gogf/vlog"
	"gorm.io/gorm"
	"io"
	"time"
)

// Callback 处理回调业务
func Callback(ctx context.Context, noticeId string) {

	// TODO: 这里可以增加分布式锁，按id进行排他

	var row = entity.Notice{NoticeId: noticeId}
	nt := row.GetNoticeTIName()

	err := dao.DB.Table(nt.TableName).
		Where("notice_id=?", row.NoticeId).
		First(&row).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		vlog.Warningf(ctx, "%s not found", noticeId)
		return
	}

	if err != nil {
		vlog.Error(ctx, err)
		return
	}

	// 状态检查
	// 0:等待被通知 1:通知成功 2:通知失败 3:正在执行通知
	if row.Status == 3 && row.RealRetryTimes > row.RetryTimes {
		// 通知失败,也达到了最大次数,就不处理
		{
			row.Status = 2
			row.StatusDesc = "超过最大重试次数"

			// 更新到数据库
			if err = dao.DB.Table(nt.TableName).Updates(&row).Error; err != nil {
				vlog.Error(ctx, err)
			}
		}
		vlog.Warningf(ctx, "status=2 and %d>=%d; skip", row.RealRetryTimes, row.RetryTimes)
		return
	}

	if row.Status == 1 {
		// 已经通知成功,也不处理
		vlog.Warningf(ctx, "status=1 skip")
		return
	}

	var desc string
	var now = time.Now()
	row.RealNoticeTime = &now
	if desc, err = SendCallback(ctx, &row); err != nil {
		// 通知失败
		row.Status = 3
		row.StatusDesc = fmt.Sprintf("通知失败: %s", err.Error())
		row.RealRetryTimes = row.RealRetryTimes + 1
	} else {
		row.Status = 1
		row.StatusDesc = fmt.Sprintf("通知成功: %s", desc)
	}

	// 更新到数据库
	if err = dao.DB.Table(nt.TableName).Updates(&row).Error; err != nil {
		vlog.Error(ctx, err)
	}

	if row.Status == 1 {
		// 如果处理成功通知,就删除timer的数据
		TimerDataSuccess <- row.NoticeId
	} else if row.Status == 3 {
		// 如果处理失败,放回队列继续处理，直到超过最大重试次数弹出
		// PoolData <- row.NoticeId
		TimerDataFailed <- &row
	}

	vlog.Infof(ctx, row.StatusDesc)
	vlog.Infof(ctx, "process notice id %s success", noticeId)
}

// SendCallback 发送回调
func SendCallback(ctx context.Context, notice *entity.Notice) (string, error) {

	switch notice.Type {
	case "http", "https":
		return SendHttpCallback(ctx, notice)
	}

	return "", fmt.Errorf("not support type: %s", notice.Type)
}

// SendHttpCallback 发送http回调
func SendHttpCallback(ctx context.Context, notice *entity.Notice) (string, error) {

	var hn = v1.HttpNotice{}
	if err := json.Unmarshal([]byte(notice.TypeData), &hn); err != nil {
		vlog.Error(ctx, err)
		return "", err
	}

	client := gclient.New().SetTimeout(time.Duration(hn.Timeout) * time.Second)
	if hn.Header != nil {
		for k, v := range hn.Header {
			client = client.SetHeader(k, v)
		}
	}
	response, err := client.Post(ctx, hn.Url, notice.Data)
	if err != nil {
		vlog.Error(ctx, err)
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 || response.StatusCode < 200 {
		vlog.Error(ctx, "http request status code: ", response.StatusCode)
		return "", err
	}

	b, err := io.ReadAll(response.Body)
	if err != nil {
		vlog.Error(ctx, err)
		return "", err
	}

	return string(b), nil
}
