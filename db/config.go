package db

import (
	"database/sql"
	"fmt"
	"go-cloud-customer/constants"
	"log"
	"os"
	"strings"

	_ "github.com/denisenkom/go-mssqldb"
)

// Config stores application configuration settings
var DB *sql.DB // Database connection pool

// InitDB loads application configuration including database connection
func InitDB() error {
	connString := mergeCredentials(constants.Cfg.ConnectionStrings.DefaultConnection)

	// Open database connection
	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Test database connection
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	log.Println("Connected to database")
	DB = db

	return nil
}

func mergeCredentials(conn string) string {
	u := os.Getenv("SQL_USERNAME")
	p := os.Getenv("SQL_PASSWORD")
	if u != "" && p != "" {
		conn = strings.ReplaceAll(conn, "[[SQL_USERNAME]]", u)
		conn = strings.ReplaceAll(conn, "[[SQL_PASSWORD]]", p)
	}
	return conn
}
