package notice

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"delay-notice/api/notice/v1"
)

func (c *ControllerV1) QueryNotice(ctx context.Context, req *v1.QueryNoticeReq) (res *v1.QueryNoticeRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
