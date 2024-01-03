// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package notice

import (
	"context"

	"delay-notice/api/notice/v1"
)

type INoticeV1 interface {
	AddNotice(ctx context.Context, req *v1.AddNoticeReq) (res *v1.AddNoticeRes, err error)
	QueryNotice(ctx context.Context, req *v1.QueryNoticeReq) (res *v1.QueryNoticeRes, err error)
}
