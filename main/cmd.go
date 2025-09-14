package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var supportedDBs = map[string]bool{
	"mysql": true,
}

func getConnection(dsn, dbType string) *sql.DB {
	if !supportedDBs[dbType] {
		log.Fatalln("UNSUPPORTED DB TYPE")
	}

	if dsn == "" {
		log.Fatalln("CONNECTION STRING CANNOT BE EMPTY")
	}

	db, err := sql.Open(dbType, dsn)
	if err != nil {
		log.Fatalf("failed to connect to %v: %v", dsn, err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping %v: %v", dsn, err)
	}

	return db
}

func getColumnNames(db *sql.DB, table string) ([]string, error) {
	query := fmt.Sprintf("SHOW COLUMNS FROM `%s`", table)

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("querying columns for %s: %w", table, err)
	}
	defer rows.Close()

	var cols []string
	for rows.Next() {
		var field, colType, null, key, defaultVal, extra sql.NullString
		if err := rows.Scan(&field, &colType, &null, &key, &defaultVal, &extra); err != nil {
			return nil, fmt.Errorf("scanning column: %w", err)
		}
		if field.Valid {
			cols = append(cols, field.String)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating columns: %w", err)
	}

	return cols, nil
}

func main() {
	leftDSN := os.Getenv("LEFT_DB_DSN")
	rightDSN := os.Getenv("RIGHT_DB_DSN")
	leftType := os.Getenv("LEFT_DB_TYPE")
	rightType := os.Getenv("RIGHT_DB_TYPE")

	leftTableName := os.Getenv("LEFT_TABLE")
	rightTableName := os.Getenv("RIGHT_TABLE")

	leftConn := getConnection(leftDSN, leftType)
	rightConn := getConnection(rightDSN, rightType)

	leftCols, err := getColumnNames(leftConn, leftTableName)
	if err != nil {
		log.Fatal(err)
	}
	rightCols, err := getColumnNames(rightConn, rightTableName)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Left table %s columns: %v\n", leftTableName, leftCols)
	log.Printf("Right table %s columns: %v\n", rightTableName, rightCols)

	diffStringArray(leftCols, rightCols)

	defer leftConn.Close()
	defer rightConn.Close()

	log.Println("Successfully connected to both databases.")
}
