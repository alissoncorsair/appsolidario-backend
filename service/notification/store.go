package notification

import (
	"database/sql"
	"time"

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
	err := row.Scan(&n.ID, &n.UserID, &n.Type, &n.ResourceID, &n.IsRead, &n.CreatedAt, &n.UpdatedAt)

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
		err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.ResourceID, &n.IsRead, &n.CreatedAt, &n.UpdatedAt)
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
	query := `INSERT INTO notifications (user_id, type, from_user_id, resource_id, is_read) VALUES ($1, $2, $3, $4, $5) RETURNING id, user_id, type, resource_id, is_read, created_at, updated_at`
	row := s.db.QueryRow(query, notification.UserID, notification.Type, notification.FromUserID, notification.ResourceID, notification.IsRead)

	return ScanRowIntoNotification(row)
}

func (s *Store) ReadNotification(notificationID int, userID int) (*types.Notification, error) {
	query := `UPDATE notifications SET is_read = true WHERE id = $1 AND user_id = $2 RETURNING id, user_id, type, resource_id, is_read, created_at, updated_at`
	row := s.db.QueryRow(query, notificationID, userID)

	notification, err := ScanRowIntoNotification(row)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return notification, nil
}

type MinimalUser struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	Email       string `json:"email"`
	UserPicture string `json:"user_picture"`
}

type NotificationResponse struct {
	ID          int         `json:"id"`
	IsRead      bool        `json:"isRead"`
	CreatedAt   time.Time   `json:"createdAt"`
	FromUser    MinimalUser `json:"fromUser"`
	Transaction struct {
		Amount float64 `json:"amount"`
	} `json:"transaction"`
}

func (s *Store) GetNotificationsByUserID(userID int) ([]NotificationResponse, error) {
	query := `
        SELECT 
            n.id, n.user_id, n.type, n.resource_id, n.is_read, n.created_at, n.updated_at,
            u.id as from_user_id, u.name, u.surname, u.email,
            pp.path as user_picture,
            t.id as transaction_id, t.amount, t.created_at as transaction_created_at
        FROM notifications n
        LEFT JOIN users u ON n.from_user_id = u.id
        LEFT JOIN profile_pictures pp ON u.id = pp.user_id
        LEFT JOIN transactions t ON n.resource_id = t.id
        WHERE n.user_id = $1 
        ORDER BY n.created_at DESC`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []NotificationResponse
	for rows.Next() {
		var detail NotificationResponse
		var notification types.Notification
		var userPicture sql.NullString
		var transactionID sql.NullString
		var transactionAmount sql.NullFloat64
		var transactionCreatedAt sql.NullTime

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.ResourceID,
			&notification.IsRead,
			&notification.CreatedAt,
			&notification.UpdatedAt,
			&detail.FromUser.ID,
			&detail.FromUser.Name,
			&detail.FromUser.Surname,
			&detail.FromUser.Email,
			&userPicture,
			&transactionID,
			&transactionAmount,
			&transactionCreatedAt,
		)
		if err != nil {
			return nil, err
		}

		detail.ID = notification.ID
		detail.IsRead = notification.IsRead
		detail.CreatedAt = notification.CreatedAt

		if userPicture.Valid {
			detail.FromUser.UserPicture = userPicture.String
		}

		if transactionAmount.Valid {
			detail.Transaction.Amount = transactionAmount.Float64
		}

		results = append(results, detail)
	}

	return results, nil
}
