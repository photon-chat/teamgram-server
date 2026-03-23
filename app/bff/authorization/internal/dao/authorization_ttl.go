package dao

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	authTTLKeyPrefix = "auth_ttl_days"
	defaultTTLDays   = 180
)

func genAuthTTLKey(userId int64) string {
	return fmt.Sprintf("%s_%d", authTTLKeyPrefix, userId)
}

func (d *Dao) SetAuthorizationTTLDays(ctx context.Context, userId int64, days int32) error {
	key := genAuthTTLKey(userId)
	if err := d.kv.SetCtx(ctx, key, strconv.Itoa(int(days))); err != nil {
		logx.WithContext(ctx).Errorf("SetAuthorizationTTLDays - kv.Set(%s) error: %v", key, err)
		return err
	}
	return nil
}
