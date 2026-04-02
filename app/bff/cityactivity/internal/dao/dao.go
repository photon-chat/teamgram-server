package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/teamgram/marmota/pkg/stores/sqlx"
	"github.com/teamgram/teamgram-server/app/bff/cityactivity/internal/config"
)

type Activity struct {
	Id               int64  `db:"id"`
	UserId           int64  `db:"user_id"`
	Title            string `db:"title"`
	Description      string `db:"description"`
	PhotoId          int64  `db:"photo_id"`
	City             string `db:"city"`
	StartTime        int64  `db:"start_time"`
	EndTime          int64  `db:"end_time"`
	MaxParticipants  int32  `db:"max_participants"`
	Status           int32  `db:"status"`
	IsGlobal         int32  `db:"is_global"`
	CreatedAt        int64  `db:"created_at"`
	UpdatedAt        int64  `db:"updated_at"`
	ParticipantCount int32  `db:"-"`
	IsJoined         bool   `db:"-"`
	CreatorName      string `db:"-"`
}

type ActivityParticipant struct {
	Id         int64  `db:"id"`
	ActivityId int64  `db:"activity_id"`
	UserId     int64  `db:"user_id"`
	City       string `db:"city"`
	JoinedAt   int64  `db:"joined_at"`
}

type Dao struct {
	db *sqlx.DB
}

func New(c config.Config) *Dao {
	return &Dao{
		db: sqlx.NewMySQL(c.Mysql),
	}
}

func (d *Dao) GetActivitiesByCity(ctx context.Context, city string, offset, limit int32) ([]*Activity, int32, error) {
	var activities []*Activity
	query := "SELECT * FROM activities WHERE (city = ? OR is_global = 1) AND status = 1 ORDER BY is_global DESC, created_at DESC LIMIT ?, ?"
	err := d.db.QueryRowsPartial(ctx, &activities, query, city, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	var count int32
	countQuery := "SELECT COUNT(*) FROM activities WHERE (city = ? OR is_global = 1) AND status = 1"
	err = d.db.QueryRow(ctx, &count, countQuery, city)
	if err != nil {
		count = int32(len(activities))
	}

	for _, a := range activities {
		a.ParticipantCount = d.getParticipantCount(ctx, a.Id)
	}

	return activities, count, nil
}

func (d *Dao) GetActivityById(ctx context.Context, id int64) (*Activity, error) {
	var activity Activity
	query := "SELECT * FROM activities WHERE id = ?"
	err := d.db.QueryRowPartial(ctx, &activity, query, id)
	if err != nil {
		return nil, err
	}
	activity.ParticipantCount = d.getParticipantCount(ctx, id)
	return &activity, nil
}

func (d *Dao) CreateActivity(ctx context.Context, a *Activity) (int64, error) {
	now := time.Now().Unix()
	a.CreatedAt = now
	a.UpdatedAt = now
	a.Status = 1

	query := "INSERT INTO activities (user_id, title, description, photo_id, city, start_time, end_time, max_participants, status, is_global, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)"
	result, err := d.db.Exec(ctx, query, a.UserId, a.Title, a.Description, a.PhotoId, a.City, a.StartTime, a.EndTime, a.MaxParticipants, a.Status, a.CreatedAt, a.UpdatedAt)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (d *Dao) EditActivity(ctx context.Context, id, userId int64, title, description string, photoId, startTime, endTime int64, status int32) error {
	now := time.Now().Unix()
	query := "UPDATE activities SET title=?, description=?, photo_id=?, start_time=?, end_time=?, status=?, updated_at=? WHERE id=? AND user_id=?"
	_, err := d.db.Exec(ctx, query, title, description, photoId, startTime, endTime, status, now, id, userId)
	return err
}

func (d *Dao) DeleteActivity(ctx context.Context, id, userId int64) error {
	now := time.Now().Unix()
	query := "UPDATE activities SET status=2, updated_at=? WHERE id=? AND user_id=?"
	_, err := d.db.Exec(ctx, query, now, id, userId)
	return err
}

func (d *Dao) JoinActivity(ctx context.Context, activityId, userId int64, city string) error {
	activity, err := d.GetActivityById(ctx, activityId)
	if err != nil {
		return err
	}

	if activity.IsGlobal == 0 && activity.City != city {
		return fmt.Errorf("city mismatch")
	}

	if activity.MaxParticipants > 0 && activity.ParticipantCount >= activity.MaxParticipants {
		return fmt.Errorf("activity is full")
	}

	now := time.Now().Unix()
	query := "INSERT IGNORE INTO activity_participants (activity_id, user_id, city, joined_at) VALUES (?, ?, ?, ?)"
	_, err = d.db.Exec(ctx, query, activityId, userId, city, now)
	return err
}

func (d *Dao) LeaveActivity(ctx context.Context, activityId, userId int64) error {
	query := "DELETE FROM activity_participants WHERE activity_id = ? AND user_id = ?"
	_, err := d.db.Exec(ctx, query, activityId, userId)
	return err
}

func (d *Dao) IsUserJoined(ctx context.Context, activityId, userId int64) bool {
	var count int32
	query := "SELECT COUNT(*) FROM activity_participants WHERE activity_id = ? AND user_id = ?"
	err := d.db.QueryRow(ctx, &count, query, activityId, userId)
	if err != nil {
		return false
	}
	return count > 0
}

func (d *Dao) getParticipantCount(ctx context.Context, activityId int64) int32 {
	var count int32
	query := "SELECT COUNT(*) FROM activity_participants WHERE activity_id = ?"
	_ = d.db.QueryRow(ctx, &count, query, activityId)
	return count
}
