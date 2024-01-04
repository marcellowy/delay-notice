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

// Cb 回调方法
func Cb(ctx context.Context, noticeId string) {

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

	var bn = &BizNotice{
		Notice: &row,
		Timer: &entity.NoticeTimer{
			NoticeId: row.NoticeId,
		},
	}

	vlog.Debug(ctx, "begin process via biz notice")
	bn.Process(ctx)
}

type BizNotice struct {
	Notice *entity.Notice
	Timer  *entity.NoticeTimer
}

// SendCallback 发送回调
func (bn *BizNotice) SendCallback(ctx context.Context) (string, error) {
	switch bn.Notice.Type {
	case "http", "https":
		return bn.SendHttpCallback(ctx)
	}
	return "", fmt.Errorf("not support type: %s", bn.Notice.Type)
}

// SendHttpCallback 发送 http 请求
func (bn *BizNotice) SendHttpCallback(ctx context.Context) (string, error) {

	var hn = v1.HttpNotice{}
	if err := json.Unmarshal([]byte(bn.Notice.TypeData), &hn); err != nil {
		vlog.Error(ctx, err)
		return "", err
	}

	client := gclient.New().SetTimeout(time.Duration(hn.Timeout) * time.Second)
	if hn.Header != nil {
		for k, v := range hn.Header {
			client = client.SetHeader(k, v)
		}
	}
	response, err := client.Post(ctx, hn.Url, bn.Notice.Data)
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

// Process 开始处理
func (bn *BizNotice) Process(ctx context.Context) {

	if bn.Notice.Status == int32(v1.NoticeStatusSuccess) {
		// 已经通知成功的数据
		vlog.Warningf(ctx, "status = 1, skip")
		return
	}

	var desc, err = bn.SendCallback(ctx)
	if err == nil {
		bn.Notice.Status = int32(v1.NoticeStatusSuccess)
		bn.Notice.StatusDesc = fmt.Sprintf("通知成功: %s", desc)

		bn.Timer.Status = 1
		_ = bn.UpdateStatus(ctx)
		return
	}

	// 通知失败
	bn.Notice.Status = int32(v1.NoticeStatusDoing)
	bn.Notice.StatusDesc = fmt.Sprintf("通知失败: %s", err.Error())
	bn.Notice.RealRetryTimes = bn.Notice.RealRetryTimes + 1

	if bn.Notice.RealRetryTimes > bn.Notice.RetryTimes {
		// 达到了最大重试次数
		bn.Notice.Status = int32(v1.NoticeStatusFailed)
		bn.Notice.StatusDesc = "超过最大重试次数"
		bn.Timer.Status = 1

		_ = bn.UpdateStatus(ctx)
		return
	}

	bn.Timer.Status = 0
	bn.Timer.PrepareNoticeTime = time.Now().Add(time.Second * time.Duration(bn.Notice.RetryInterval))

	_ = bn.UpdateStatus(ctx)
	return
}

// UpdateStatus 事务更新状态
func (bn *BizNotice) UpdateStatus(ctx context.Context) error {

	nt := bn.Notice.GetNoticeTIName()

	return dao.DB.Transaction(func(tx *gorm.DB) (err error) {

		err = tx.Table(nt.TableName).Where("notice_id = ?", bn.Notice.NoticeId).Updates(map[string]interface{}{
			"status":           bn.Notice.Status,
			"status_desc":      bn.Notice.StatusDesc,
			"real_retry_times": bn.Notice.RealRetryTimes,
		}).Error

		if err != nil {
			vlog.Error(ctx, err)
			return err
		}

		if bn.Timer.Status == 1 {
			err = tx.Unscoped().Where("notice_id = ?", bn.Notice.NoticeId).Delete(&entity.NoticeTimer{}).Error
		} else {
			err = tx.Model(&entity.NoticeTimer{}).Where("notice_id = ?", bn.Notice.NoticeId).Updates(map[string]interface{}{
				"status":              bn.Timer.Status,
				"prepare_notice_time": bn.Timer.PrepareNoticeTime,
			}).Error
		}

		if nil != err {
			vlog.Error(ctx, err)
			return err
		}

		return nil
	})
}
