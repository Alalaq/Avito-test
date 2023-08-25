package services

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type UserService struct {
	db *sql.DB // Database connection
}

type SegmentHistoryEntry struct {
	UserID      int
	SegmentName string
	Operation   string
	SegmentTime time.Time
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

// CreateUser @Summary Create User
// @Tags users
// @Description Create a new user
// @Produce json
// @Success 200 {integer} int "User ID"
func (u *UserService) CreateUser() (int, error) {
	result, err := u.db.Exec("INSERT INTO users (created_at) VALUES (CURRENT_TIMESTAMP)")
	if err != nil {
		return 0, err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(userID), nil
}

// AddUserToSegment @Summary Add User to Segment
// @Tags users
// @Description Add user to a segment
// @Accept json
// @Produce json
// @Param user_id body int true "User ID"
// @Param segments_to_add body []string true "Segments to add"
// @Param segments_to_remove body []string false "Segments to remove"
// @Param expires_at body string true "Expiry timestamp (RFC3339 format)"
// @Success 200 {string} string "Success message"
func (u *UserService) AddUserToSegment(userID int, segmentID int, expiresAt time.Time) error {
	_, err := u.db.Exec("INSERT INTO user_segments (user_id, segment_id, expires_at) VALUES (?, ?, ?)", userID, segmentID, expiresAt)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserService) RemoveUserFromSegment(userID int, segmentID int) error {
	_, err := u.db.Exec("DELETE FROM user_segments WHERE user_id = ? AND segment_id = ?", userID, segmentID)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserService) GetAllUserIDs() ([]int, error) {
	query := "SELECT id FROM users"
	rows, err := u.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []int
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

func (u *UserService) LogSegmentHistory(userID int, segmentID int, operation string, timestamp time.Time) error {
	_, err := u.db.Exec("INSERT INTO segment_history (user_id, segment_id, operation, timestamp) VALUES (?, ?, ?, ?)", userID, segmentID, operation, timestamp)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserService) GetSegmentHistoryByPeriod(year, month int) ([]SegmentHistoryEntry, error) {
	// Define a struct to hold segment history entries

	var segmentHistory []SegmentHistoryEntry

	query := `
		SELECT user_id, segments.slug, operation, timestamp
		FROM segment_history
		JOIN segments ON segment_history.segment_id = segments.id
		WHERE YEAR(timestamp) = ? AND MONTH(timestamp) = ?
	`
	rows, err := u.db.Query(query, year, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var entry SegmentHistoryEntry
		var timestampStr string // Declare a string to hold the timestamp as string

		if err := rows.Scan(&entry.UserID, &entry.SegmentName, &entry.Operation, &timestampStr); err != nil {
			return nil, err
		}

		entry.SegmentTime, err = time.Parse("2006-01-02 15:04:05", timestampStr) // Parse the timestamp string
		if err != nil {
			return nil, err
		}

		segmentHistory = append(segmentHistory, entry)
	}

	return segmentHistory, nil
}

func (u *UserService) AddUserToSegments(userID int, segmentIDsToAdd []int, segmentIDsToRemove []int, expiresAt time.Time) error {
	tx, err := u.db.Begin()
	if err != nil {
		return err
	}

	for _, segmentToAdd := range segmentIDsToAdd {
		_, err = tx.Exec("INSERT INTO user_segments (user_id, segment_id, expires_at) VALUES (?, ?, ?)", userID, segmentToAdd, expiresAt)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	for _, segmentToRemove := range segmentIDsToRemove {
		_, err = tx.Exec("DELETE FROM user_segments WHERE user_id = ? AND segment_id = ?", userID, segmentToRemove)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
