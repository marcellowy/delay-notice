package notice

import (
	"context"
	"delay-notice/internal/dao"
	"delay-notice/internal/model/entity"
	"encoding/json"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/marcellowy/go-common/gogf/vlog"
	"gorm.io/gorm"
	"time"

	"delay-notice/api/notice/v1"
)

func (c *ControllerV1) AddNotice(ctx context.Context, req *v1.AddNoticeReq) (res *v1.AddNoticeRes, err error) {
	res = &v1.AddNoticeRes{}
	if req.Type == "" || req.Delay < 0 || req.RetryTimes < 0 || req.RetryTimes > 10 {
		vlog.Error(ctx, gcode.CodeInvalidParameter)
		return nil, gerror.NewCode(gcode.CodeInvalidParameter)
	}

	httpBytes, err := json.Marshal(req.Http)
	if err != nil {
		vlog.Error(ctx, err)
		return nil, gerror.NewCode(gcode.CodeInvalidParameter)
	}

	var prepareNoticeTime = time.Now().Add(time.Second * time.Duration(req.Delay))

	modelNotice := entity.Notice{
		Type:              req.Type,
		TypeData:          string(httpBytes),
		Delay:             req.Delay,
		PrepareNoticeTime: &prepareNoticeTime,
		Status:            0,
		StatusDesc:        "等待处理",
		RetryTimes:        req.RetryTimes,
		RetryInterval:     req.RetryInterval,
		RealRetryTimes:    0,
		Data:              req.Data,
	}

	// 创建id,确定表名
	nt := modelNotice.Generate()

	modelNoticeTimer := entity.NoticeTimer{
		NoticeId:          nt.NoticeId,
		PrepareNoticeTime: prepareNoticeTime,
		Status:            0,
	}

	// 事务写入
	err = dao.DB.Transaction(func(tx *gorm.DB) error {
		var e error
		if e = tx.Table(nt.TableName).Create(&modelNotice).Error; e != nil {
			vlog.Error(ctx, e)
			return e
		}

		if e = tx.Create(&modelNoticeTimer).Error; e != nil {
			vlog.Error(ctx, err)
			return e
		}

		return nil
	})

	if err != nil {
		vlog.Error(ctx, err)
		return nil, gerror.NewCode(gcode.CodeDbOperationError)
	}

	res.NoticeId = nt.NoticeId

	return res, nil
}
