package notice

import (
	"context"
	"delay-notice/internal/dao"
	"delay-notice/internal/model/entity"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/marcellowy/go-common/gogf/vlog"

	"delay-notice/api/notice/v1"
)

func (c *ControllerV1) QueryNotice(ctx context.Context, req *v1.QueryNoticeReq) (res *v1.QueryNoticeRes, err error) {
	res = &v1.QueryNoticeRes{}
	var noticeRow = entity.Notice{}
	if err = dao.DB.First(&noticeRow).Error; err != nil {
		vlog.Error(ctx, err)
		return nil, gerror.NewCode(gcode.CodeDbOperationError)
	}
	res.Status = v1.NoticeStatus(noticeRow.Status)
	res.Message = noticeRow.StatusDesc
	return
}
