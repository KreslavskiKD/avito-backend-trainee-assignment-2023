package db_postgresql

import (
	"database/sql"
	"fmt"
	"os"
	"service_api/cmd/internal/db"
	"time"

	"github.com/lib/pq"

	_ "github.com/lib/pq"
)

type PostgreSqlStorage struct {
	db *sql.DB
}

func New(storagePath string) (*PostgreSqlStorage, error) {
	const errorPath = "db.db_postgresql.New"

	db, err := sql.Open("postgres", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errorPath, err)
	}

	// users table
	stmt, err := db.Prepare(
		`CREATE TABLE IF NOT EXISTS users (
    		id serial PRIMARY KEY, 
    		name varchar(255) NOT NULL
  		);`,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", errorPath, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errorPath, err)
	}

	// segments table
	// Basically we have 2 primary keys
	stmt, err = db.Prepare(
		`CREATE TABLE IF NOT EXISTS segments (
    		id serial PRIMARY KEY,
    		name varchar(255) UNIQUE NOT NULL		
  		);`,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", errorPath, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errorPath, err)
	}

	// user_segments table
	stmt, err = db.Prepare(
		`CREATE TABLE IF NOT EXISTS user_segments (
    		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    		segment_id INTEGER REFERENCES segments(id) ON DELETE CASCADE,
    		PRIMARY KEY (user_id, segment_id)
  		);`,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", errorPath, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errorPath, err)
	}

	// history table
	stmt, err = db.Prepare(
		`CREATE TABLE IF NOT EXISTS history (
    		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    		segment_id INTEGER REFERENCES segments(id) ON DELETE CASCADE,
    		operation VARCHAR(10),
    		timestamp TIMESTAMP NOT NULL DEFAULT NOW()
		);`,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", errorPath, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errorPath, err)
	}

	return &PostgreSqlStorage{db: db}, nil
}

func (storage *PostgreSqlStorage) CreateSegment(name string) error {
	const errorPath = "db.db_postgresql.CreateSegment"

	stmt, err := storage.db.Prepare("INSERT INTO segments(name) VALUES($1)")

	if err != nil {
		return fmt.Errorf("%s: %w", errorPath, err)
	}

	_, err = stmt.Exec(name)
	if err != nil {
		if postgresErr, ok := err.(*pq.Error); ok && postgresErr.Code == AlreadyExistsError {
			return fmt.Errorf("%s: %w", errorPath, db.ErrSegmentAlreadyExists)
		}
		return fmt.Errorf("%s: %w", errorPath, err)
	}

	return nil
}

func (storage *PostgreSqlStorage) DeleteSegment(name string) error {
	const errorPath = "db.db_postgresql.DeleteSegment"
	stmt, err := storage.db.Prepare(
		`DELETE FROM segments 
		WHERE name = $1;`,
	)

	if err != nil {
		return fmt.Errorf("%s: %w", errorPath, err)
	}

	_, err = stmt.Exec(name)
	if err != nil {
		return fmt.Errorf("%s: %w", errorPath, err)
	}
	return nil
}

func (storage *PostgreSqlStorage) RemoveSegmentsFromUser(removeSegments []string, userId int64) error {
	const errorPath = "db.db_postgresql.RemoveSegmentsFromUser"

	stmt, err := storage.db.Prepare(
		`DELETE FROM user_segments 
		WHERE user_id = $1 AND segment_id = ANY(
			SELECT id FROM segments 
			WHERE name = ANY($2)
		);`,
	)

	if err != nil {
		return fmt.Errorf("%s: %w", errorPath, err)
	}

	_, err = stmt.Exec(userId, pq.Array(removeSegments))
	if err != nil {
		return fmt.Errorf("%s: %w", errorPath, err)
	}

	stmt, err = storage.db.Prepare(
		`INSERT INTO history(user_id, segment_id,operation) 
		SELECT $1, id, 'Removed' 
		FROM segments 
		WHERE name = ANY($2);`,
	)

	if err != nil {
		return fmt.Errorf("%s: %w", errorPath, err)
	}

	_, err = stmt.Exec(userId, pq.Array(removeSegments))
	if err != nil {
		if postgresErr, ok := err.(*pq.Error); ok && postgresErr.Code == NotFoundEror {
			return fmt.Errorf("%s: %w", errorPath, db.ErrUserNotFound)
		}
		return fmt.Errorf("%s: %w", errorPath, err)
	}
	return nil
}

func (storage *PostgreSqlStorage) AddSegmentsToUser(addSegments []string, userId int64) error {
	const errorPath = "db.db_postgresql.RemoveSegmentsFromUser"

	stmt, err := storage.db.Prepare(
		`INSERT INTO user_segments(user_id, segment_id) 
		SELECT $1, id 
		FROM segments 
		WHERE name = ANY($2);`,
	)

	if err != nil {
		return fmt.Errorf("%s: %w", errorPath, err)
	}

	_, err = stmt.Exec(userId, pq.Array(addSegments))
	if err != nil {
		if postgresErr, ok := err.(*pq.Error); ok && postgresErr.Code == AlreadyExistsError {
			return fmt.Errorf("%s: %w", errorPath, db.ErrUserAlreadyHasSegment)
		}
		if postgresErr, ok := err.(*pq.Error); ok && postgresErr.Code == NotFoundEror {
			return fmt.Errorf("%s: %w", errorPath, db.ErrUserNotFound)
		}
		return fmt.Errorf("%s: %w", errorPath, err)
	}

	stmt, err = storage.db.Prepare(`
		INSERT INTO history(user_id, segment_id,operation) 
		SELECT $1, id, 'Added' 
		FROM segments 
		WHERE name = ANY($2);`)

	if err != nil {
		return fmt.Errorf("%s: %w", errorPath, err)
	}
	_, err = stmt.Exec(userId, pq.Array(addSegments))
	if err != nil {
		return fmt.Errorf("%s: %w", errorPath, err)
	}
	return nil
}

func (storage *PostgreSqlStorage) GetSegments(userId int64) ([]string, error) {
	const errorPath = "db.db_postgresql.GetSegments"

	stmt, err := storage.db.Prepare(
		`SELECT name 
		FROM segments 
		JOIN user_segments ON segments.id = user_segments.segment_id 
			AND user_id = $1;`,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", errorPath, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(userId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errorPath, err)
	}
	defer rows.Close()

	var segments []string

	for rows.Next() {
		var segmentName string
		if err := rows.Scan(&segmentName); err != nil {
			return nil, fmt.Errorf("%s: %w", errorPath, err)
		}
		segments = append(segments, segmentName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", errorPath, err)
	}

	return segments, nil
}

func (storage *PostgreSqlStorage) ReassignSegments(addSegments []string, removeSegments []string, userId int64) error {
	const errorPath = "storage.postgres.ReassignSegments"

	err := storage.RemoveSegmentsFromUser(removeSegments, userId)
	if err != nil {
		return fmt.Errorf("%s: %w", errorPath, err)
	}
	err = storage.AddSegmentsToUser(addSegments, userId)
	if err != nil {
		return fmt.Errorf("%s: %w", errorPath, err)
	}
	return nil
}

func (storage *PostgreSqlStorage) PrepareUserHistoryFile(userId int64, year int, month time.Month) error {
	startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	rows, err := storage.db.Query(`
		SELECT user_id, segments.name, operation, timestamp 
		FROM history 
		JOIN segments ON segment_id = segments.id 
		WHERE timestamp >= $1 AND timestamp <= $2;`,
		startOfMonth,
		endOfMonth,
	)

	if err != nil {
		return err
	}
	defer rows.Close()

	filename := fmt.Sprint(userId) + "_history_report.csv"

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintln(file, "user_id;segment;operation;date and time")

	for rows.Next() {
		var userID int
		var segmentName string
		var operation string
		var timestamp time.Time

		err := rows.Scan(&userID, &segmentName, &operation, &timestamp)
		if err != nil {
			return err
		}
		fmt.Fprintf(file, "%d;%s;%s;%s\n", userID, segmentName, operation, timestamp.Format("2006-01-02 15:04:05"))
	}

	return nil
}
