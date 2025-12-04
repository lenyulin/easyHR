package oss

import (
	"context"
)

type OSSHandler interface {
	Upload(ctx context.Context, fileDir string) (string, string, error)
	Find(ctx context.Context, uid int64) error
	Delete(ctx context.Context, uid int64) error
}
