package services

import (
	"avitoGoProject/services"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type APIHandlers struct {
	userService    *services.UserService
	segmentService *services.SegmentService
}

func NewAPIHandlers(userService *services.UserService, segmentService *services.SegmentService) *APIHandlers {
	return &APIHandlers{
		userService:    userService,
		segmentService: segmentService,
	}
}

// CreateUserHandler @Summary Create a new user
// @Description Create a new user and return the user ID.
// @Tags users
// @Produce json
// @Success 200 {object} map[string]int "User ID"
// @Failure 500 {string} string "Internal Server Error"
// @Router /users/create [post]
func (a *APIHandlers) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := a.userService.CreateUser()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]int{"user_id": userID}
	jsonResponse(w, response)
}

// UpdateUserSegmentsHandler @Summary Update user segments
// @Description Update user segments by adding or removing specified segments.
// @Tags users
// @Accept json
// @Produce json
// @Param user_id body int true "User ID"
// @Param segments_to_add body array true "Segments to add"
// @Param segments_to_remove body array true "Segments to remove"
// @Param expires_at body string true "Expiration date"
// @Success 200 {object} map[string]interface{} "Response message"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /users/update-segments [post]
func (a *APIHandlers) UpdateUserSegmentsHandler(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		UserID           int      `json:"user_id"`
		SegmentsToAdd    []string `json:"segments_to_add"`
		SegmentsToRemove []string `json:"segments_to_remove"`
		ExpiresAt        string   `json:"expires_at"` // Expects a string representation of a valid datetime
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Convert the ExpiresAt string to a time.Time value
	expiresAt, err := time.Parse(time.RFC3339, requestData.ExpiresAt)
	if err != nil {
		http.Error(w, "Invalid datetime format for expires_at", http.StatusBadRequest)
		return
	}

	var responseMessage []string

	for _, segmentSlugToAdd := range requestData.SegmentsToAdd {
		segmentID, err := a.segmentService.GetSegmentIDBySlug(segmentSlugToAdd)
		if err != nil {
			responseMessage = append(responseMessage, fmt.Sprintf(`"%s" doesn't exist`, segmentSlugToAdd))
			continue
		}
		// Add the user to the segment and log the operation
		err = a.userService.AddUserToSegments(requestData.UserID, []int{segmentID}, nil, expiresAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = a.userService.LogSegmentHistory(requestData.UserID, segmentID, "add", time.Now())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		responseMessage = append(responseMessage, fmt.Sprintf(`"%s" added successfully`, segmentSlugToAdd))
	}

	for _, segmentSlugToRemove := range requestData.SegmentsToRemove {
		segmentID, err := a.segmentService.GetSegmentIDBySlug(segmentSlugToRemove)
		if err != nil {
			responseMessage = append(responseMessage, fmt.Sprintf(`"%s" doesn't exist`, segmentSlugToRemove))
			continue
		}
		// Check if the user is linked to the segment before removal
		isLinked, err := a.userService.IsUserLinkedToSegment(requestData.UserID, segmentID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !isLinked {
			responseMessage = append(responseMessage, fmt.Sprintf(`"%s" is not linked to the user`, segmentSlugToRemove))
			continue
		}
		// Remove the user from the segment and log the operation
		err = a.userService.AddUserToSegments(requestData.UserID, nil, []int{segmentID}, expiresAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = a.userService.LogSegmentHistory(requestData.UserID, segmentID, "remove", time.Now())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		responseMessage = append(responseMessage, fmt.Sprintf(`"%s" removed successfully`, segmentSlugToRemove))
	}

	jsonResponse(w, map[string]interface{}{"message": responseMessage})
}

// CreateSegmentHandler @Summary Create a new segment
// @Description Create a new segment with auto-add option and return success message.
// @Tags segments
// @Accept json
// @Produce json
// @Param slug body string true "Slug of the segment"
// @Param auto_add body bool true "Auto Add flag"
// @Param auto_pct body int true "Auto Percentage"
// @Success 200 {object} map[string]string "Response message"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /segments/create [post]
func (a *APIHandlers) CreateSegmentHandler(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		Slug    string `json:"slug"`
		AutoAdd bool   `json:"auto_add"`
		AutoPct int    `json:"auto_pct"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	segmentID, err := a.segmentService.CreateSegmentAndGetID(requestData.Slug, requestData.AutoAdd, requestData.AutoPct)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if requestData.AutoAdd {
		// Get all user IDs and shuffle them
		userIDs, err := a.userService.GetAllUserIDs()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		rand.Shuffle(len(userIDs), func(i, j int) { userIDs[i], userIDs[j] = userIDs[j], userIDs[i] })

		// Calculate the number of users to add based on percentage
		numUsersToAdd := (len(userIDs) * requestData.AutoPct) / 100

		// Add the calculated number of users to the segment
		expiresAt := time.Now().Add(time.Duration(requestData.AutoPct) * 24 * time.Hour) // Calculate the expiration time
		for i := 0; i < numUsersToAdd; i++ {
			err := a.userService.AddUserToSegment(userIDs[i], segmentID, expiresAt)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	jsonResponse(w, map[string]string{"message": "Segment created"})
}

// DeleteSegmentHandler @Summary Delete a segment
// @Description Delete a segment by slug and return success message.
// @Tags segments
// @Produce json
// @Param slug query string true "Slug of the segment"
// @Success 200 {object} map[string]string "Response message"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /segments/delete [delete]
func (a *APIHandlers) DeleteSegmentHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		http.Error(w, "Missing 'slug' parameter", http.StatusBadRequest)
		return
	}

	err := a.segmentService.DeleteSegment(slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]string{"message": "Segment deleted"})
}

// GetUserSegmentsHandler @Summary Get user's segments
// @Description Get a list of segments linked to a user by providing the user ID.
// @Tags segments
// @Produce json
// @Param user_id query int true "User ID"
// @Success 200 {object} map[string][]string "Segments"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /segments/user-segments [get]
func (a *APIHandlers) GetUserSegmentsHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid 'user_id' parameter", http.StatusBadRequest)
		return
	}

	segments, err := a.segmentService.GetUserSegments(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string][]string{"segments": segments})
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(data)
}

// GenerateSegmentHistoryReportHandler @Summary Generate segment history report
// @Description Generate a CSV report of segment history for a specified year and month.
// @Tags segments
// @Produce plain
// @Param year query int true "Year"
// @Param month query int true "Month"
// @Success 200 {string} plain "CSV report"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /segments/history-report [get]
func (a *APIHandlers) GenerateSegmentHistoryReportHandler(w http.ResponseWriter, r *http.Request) {
	// Parse year and month from request parameters
	year, err := strconv.Atoi(r.FormValue("year"))
	if err != nil {
		http.Error(w, "Invalid year format", http.StatusBadRequest)
		return
	}
	month, err := strconv.Atoi(r.FormValue("month"))
	if err != nil {
		http.Error(w, "Invalid month format", http.StatusBadRequest)
		return
	}

	// Fetch segment history for the specified period
	segmentHistory, err := a.userService.GetSegmentHistoryByPeriod(year, month)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate CSV content
	var csvContent bytes.Buffer
	csvWriter := csv.NewWriter(&csvContent)
	for _, entry := range segmentHistory {
		csvWriter.Write([]string{
			strconv.Itoa(entry.UserID),
			entry.SegmentName,
			entry.Operation,
			entry.SegmentTime.Format(time.RFC3339),
		})
	}
	csvWriter.Flush()

	// Set response headers
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=segment_history.csv")

	// Write CSV content to the response
	_, _ = w.Write(csvContent.Bytes())
}
