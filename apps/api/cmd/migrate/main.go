package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	dsn := os.Getenv("GATEWARDEN_DB_DSN")
	if dsn == "" {
		dsn = "postgres://gatewarden:gatewarden@localhost:5432/gatewarden?sslmode=disable&connect_timeout=10"
	}

	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"up"}
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	migrationsDir := findMigrationsDir()
	if err := goose.Run("status", db, migrationsDir); err != nil {
		log.Printf("status: %v", err)
	}

	if err := goose.Run(args[0], db, migrationsDir, args[1:]...); err != nil {
		log.Fatalf("migrate %s: %v", args[0], err)
	}
}

func findMigrationsDir() string {
	candidates := []string{
		"/migrations",
		"../../migrations",
		"../../../migrations",
		"migrations",
	}
	for _, d := range candidates {
		abs, _ := filepath.Abs(d)
		info, err := os.Stat(abs)
		if err != nil || !info.IsDir() {
			continue
		}
		entries, err := os.ReadDir(abs)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				return abs
			}
		}
	}
	log.Fatal("no migrations directory with SQL files found")
	return ""
}
