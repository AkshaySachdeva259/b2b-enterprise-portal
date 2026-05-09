package main

import (
	"log"
	"net/http"

	"com.jetapcglobal.b2b.com/database"
	"com.jetapcglobal.b2b.com/router"
)

func main() {
	db := database.Connect()
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB: %v", err)
	}
	defer sqlDB.Close()

	handler := router.New(db)

	srv := &http.Server{
		Addr:    ":3080",
		Handler: handler,
	}

	log.Println("server listening on :3080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
