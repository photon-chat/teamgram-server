package dao

import (
	"context"
)

type userPasswordDO struct {
	PasswordHash string `db:"password_hash"`
}

type userPasswordByPhoneDO struct {
	UserId       int64  `db:"user_id"`
	PasswordHash string `db:"password_hash"`
}

// SaveUserPassword 存储用户密码哈希
func (d *Dao) SaveUserPassword(ctx context.Context, userId int64, passwordHash string) error {
	query := "INSERT INTO user_passwords (user_id, password_hash) VALUES (?, ?) ON DUPLICATE KEY UPDATE password_hash = VALUES(password_hash), updated_at = NOW()"
	_, err := d.DB.Exec(ctx, query, userId, passwordHash)
	return err
}

// GetUserPassword 获取用户密码哈希
func (d *Dao) GetUserPassword(ctx context.Context, userId int64) (string, error) {
	var do userPasswordDO
	query := "SELECT password_hash FROM user_passwords WHERE user_id = ? AND deleted = 0"
	err := d.DB.QueryRowPartial(ctx, &do, query, userId)
	if err != nil {
		return "", err
	}
	return do.PasswordHash, nil
}

// GetUserPasswordByPhone 通过手机号获取 user_id 和密码哈希
func (d *Dao) GetUserPasswordByPhone(ctx context.Context, phone string) (int64, string, error) {
	var do userPasswordByPhoneDO
	query := "SELECT u.id AS user_id, p.password_hash FROM users u JOIN user_passwords p ON u.id = p.user_id WHERE u.phone = ? AND u.deleted = 0 AND p.deleted = 0"
	err := d.DB.QueryRowPartial(ctx, &do, query, phone)
	if err != nil {
		return 0, "", err
	}
	return do.UserId, do.PasswordHash, nil
}
