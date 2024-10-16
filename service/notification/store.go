package notification

import (
	"database/sql"

	"github.com/alissoncorsair/appsolidario-backend/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func ScanRowIntoNotification(row *sql.Row) (*types.Notification, error) {
	var n types.Notification
	err := row.Scan(&n.ID, &n.UserID, &n.Type, &n.ResourceID, &n.IsRead)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &n, nil
}

func ScanRowsIntoNotifications(rows *sql.Rows) ([]*types.Notification, error) {
	var notifications []*types.Notification

	for rows.Next() {
		var n types.Notification
		err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.ResourceID, &n.IsRead)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, &n)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (s *Store) CreateNotification(notification *types.Notification) (*types.Notification, error) {
	query := `INSERT INTO notifications (user_id, type, from_user_id, resource_id, is_read) VALUES ($1, $2, $3, $4, $5) RETURNING id, user_id, type, resource_id, is_read`
	row := s.db.QueryRow(query, notification.UserID, notification.Type, notification.FromUserID, notification.ResourceID, notification.IsRead)

	return ScanRowIntoNotification(row)
}

func (s *Store) ReadNotification(notificationID int) (*types.Notification, error) {
	query := `UPDATE notifications SET is_read = true WHERE id = $1 RETURNING id, user_id, type, resource_id, is_read`
	row := s.db.QueryRow(query, notificationID)

	return ScanRowIntoNotification(row)
}

func (s *Store) GetNotificationsByUserID(userID int) ([]*types.Notification, error) {
	query := `SELECT id, user_id, type, resource_id, is_read FROM notifications WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return ScanRowsIntoNotifications(rows)
}
