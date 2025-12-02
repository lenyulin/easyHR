package retry

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// Retry 带退避策略的重试函数
// ctx: 上下文
// maxAttempts: 最大重试次数
// interval: 初始重试间隔
// fn: 要重试的函数（返回error则重试，返回nil则成功）
func Retry(ctx context.Context, maxAttempts int, interval time.Duration, fn func() error) error {
	// 配置退避策略（指数退避）
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = interval
	bo.MaxElapsedTime = 0 // 不限制总时间

	// 限制最大重试次数
	bo = backoff.WithMaxRetries(bo, uint64(maxAttempts-1))
	// 绑定上下文（支持取消）
	bo = backoff.WithContext(bo, ctx)

	// 执行重试
	return backoff.Retry(fn, bo)
}
