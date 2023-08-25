package main

import (
	handlers "avitoGoProject/handlers"
	"avitoGoProject/services"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
)

func main() {
	// Initialize database connection
	db, err := sql.Open("mysql", "root:12345@tcp(localhost:3306)/avito_project_db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize services
	userService := services.NewUserService(db)
	segmentService := services.NewSegmentService(db)

	apiHandlers := handlers.NewAPIHandlers(userService, segmentService)

	// Set up HTTP routes
	router := http.NewServeMux()
	router.HandleFunc("/users/create", allowOnly(apiHandlers.CreateUserHandler, http.MethodPost))
	router.HandleFunc("/users/update-segments", allowOnly(apiHandlers.UpdateUserSegmentsHandler, http.MethodPost))
	router.HandleFunc("/users/history-report", allowOnly(apiHandlers.GenerateSegmentHistoryReportHandler, http.MethodGet))
	router.HandleFunc("/segments/create", allowOnly(apiHandlers.CreateSegmentHandler, http.MethodPost))
	router.HandleFunc("/segments/delete", allowOnly(apiHandlers.DeleteSegmentHandler, http.MethodDelete))
	router.HandleFunc("/segments/user-segments", allowOnly(apiHandlers.GetUserSegmentsHandler, http.MethodGet))
	router.Handle("/swagger/", httpSwagger.WrapHandler)

	// Start the HTTP server
	serverAddr := "localhost:8080"
	fmt.Printf("Server is listening on %s...\n", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, router))
}

func allowOnly(next http.HandlerFunc, method string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "Method not allowed. Only "+method+" method is allowed.", http.StatusMethodNotAllowed)
			return
		}
		next(w, r)
	}
}
