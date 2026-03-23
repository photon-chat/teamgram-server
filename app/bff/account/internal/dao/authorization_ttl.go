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

func (d *Dao) GetAuthorizationTTLDays(ctx context.Context, userId int64) int32 {
	if d.kv == nil {
		return int32(defaultTTLDays)
	}

	key := genAuthTTLKey(userId)
	val, err := d.kv.GetCtx(ctx, key)
	if err != nil {
		logx.WithContext(ctx).Errorf("GetAuthorizationTTLDays - kv.Get(%s) error: %v", key, err)
		return int32(defaultTTLDays)
	}
	if val == "" {
		return int32(defaultTTLDays)
	}

	days, err := strconv.Atoi(val)
	if err != nil {
		logx.WithContext(ctx).Errorf("GetAuthorizationTTLDays - strconv.Atoi(%s) error: %v", val, err)
		return int32(defaultTTLDays)
	}
	return int32(days)
}

func (d *Dao) SetAuthorizationTTLDays(ctx context.Context, userId int64, days int32) error {
	if d.kv == nil {
		return nil
	}

	key := genAuthTTLKey(userId)
	return d.kv.SetCtx(ctx, key, strconv.Itoa(int(days)))
}
