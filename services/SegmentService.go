package services

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql" // MySQL driver import
)

type SegmentService struct {
	db *sql.DB // Database connection
}

func NewSegmentService(db *sql.DB) *SegmentService {
	return &SegmentService{db: db}
}

// CreateSegmentAndGetID @Summary Create a new segment and get its ID
// @Description Create a new segment and retrieve its ID by providing slug, autoAdd, and autoPct.
// @Tags segments
// @Accept json
// @Produce json
// @Param slug body string true "Slug of the segment"
// @Param autoAdd body bool true "Auto Add flag"
// @Param autoPct body int true "Auto Percentage"
// @Success 200 {integer} int "Segment ID"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
func (s *SegmentService) CreateSegmentAndGetID(slug string, autoAdd bool, autoPct int) (int, error) {
	query := "INSERT INTO segments (slug, auto_add, auto_pct, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)"
	_, err := s.db.Exec(query, slug, autoAdd, autoPct)
	if err != nil {
		return 0, err
	}

	// Get the ID of the newly inserted segment
	segmentID, err := s.GetSegmentIDBySlug(slug)
	if err != nil {
		return 0, err
	}

	return segmentID, nil
}

// DeleteSegment @Summary Delete a segment by slug
// @Description Delete a segment by providing its slug.
// @Tags segments
// @Accept json
// @Produce json
// @Param slug path string true "Slug of the segment"
// @Success 200 {string} string "Segment deleted"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
func (s *SegmentService) DeleteSegment(slug string) error {
	query := "DELETE FROM segments WHERE slug = ?"
	_, err := s.db.Exec(query, slug)
	if err != nil {
		return err
	}
	return nil
}

// GetSegmentIDBySlug @Summary Get segment ID by slug
// @Description Get segment ID by providing its slug.
// @Tags segments
// @Accept json
// @Produce json
// @Param slug path string true "Slug of the segment"
// @Success 200 {integer} int "Segment ID"
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Segment not found"
// @Failure 500 {string} string "Internal Server Error"
func (s *SegmentService) GetSegmentIDBySlug(slug string) (int, error) {
	var segmentID int

	query := "SELECT id FROM segments WHERE slug = ?"
	err := s.db.QueryRow(query, slug).Scan(&segmentID)
	if err != nil {
		return 0, err
	}

	return segmentID, nil
}

// GetUserSegments @Summary Get user's segments by user ID
// @Description Get a list of segments linked to a user by providing the user ID.
// @Tags segments
// @Accept json
// @Produce json
// @Param userID path int true "User ID"
// @Success 200 {array} string "List of segment slugs"
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Internal Server Error"
func (s *SegmentService) GetUserSegments(userID int) ([]string, error) {
	var segments []string

	query := `
		SELECT segments.slug
		FROM segments
		JOIN user_segments ON segments.id = user_segments.segment_id
		WHERE user_segments.user_id = ?
	`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var segment string
		if err := rows.Scan(&segment); err != nil {
			return nil, err
		}
		segments = append(segments, segment)
	}

	return segments, nil
}

func (u *UserService) IsUserLinkedToSegment(userID int, segmentID int) (bool, error) {
	query := "SELECT COUNT(*) FROM user_segments WHERE user_id = ? AND segment_id = ?"
	var count int
	err := u.db.QueryRow(query, userID, segmentID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
