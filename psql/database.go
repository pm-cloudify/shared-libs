package database

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// users table data
type User struct {
	ID             int
	Username       string
	HashedPassword string
	CreatedAt      string
}

// uploads table data
type Upload struct {
	ID        int
	Filename  string
	Username  string
	CreatedAt string
	Enable    bool
	FileKey   string
}

// database
var DB *pgxpool.Pool

// initialize database connection
func InitDB(connectionStr string) error {
	config, err := pgxpool.ParseConfig(connectionStr)
	if err != nil {
		return err
	}

	DB, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return err
	}

	log.Println("Successfully connected to the database")

	return nil
}

// close database connection
func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("Database connection closed")
	}
}

// TODO: add user to users table
func AddUser(username, hashed_password string) error {
	query := `INSERT INTO users(username, password_hash) VALUES ($1, $2)`

	if _, err := DB.Exec(context.Background(), query, username, hashed_password); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return errors.New("username exists")
			}
		}
		log.Println(pgErr)

		return err
	}
	return nil
}

// get user data by username from users table
func GetUserByUsername(username string) (*User, error) {
	// TODO: check for query injection ?
	query := `SELECT id, username, password_hash FROM users WHERE username = $1`

	row := DB.QueryRow(context.Background(), query, username)

	var user User

	err := row.Scan(&user.ID, &user.Username, &user.HashedPassword)
	if err != nil && err.Error() == "no rows in result set" {
		log.Println(err.Error())
		return nil, nil
	} else if err != nil {
		log.Printf("Failed to fetch user: %s\n", err.Error())
		return nil, err
	}

	return &user, nil
}

// add a upload info
func AddUploadedFileInfo(filename, username, fileKey string) error {
	query := `INSERT INTO uploads(filename, username, file_key) VALUES ($1, $2, $3)`

	if _, err := DB.Query(context.Background(), query, filename, username, fileKey); err != nil {
		return err
	}
	return nil
}

// TODO: Paginate data
// TODO: Config limit rate
// retrieve data on uploads table
func GetUploadsByUsername(username string) ([]Upload, error) {
	query := `SELECT id, filename, username, enable FROM uploads WHERE username = $1 LIMIT 5`

	rows, err := DB.Query(context.Background(), query, username)
	if err != nil {
		return nil, err
	}

	var uploads []Upload

	// TODO: better implementation with CollectRows
	for rows.Next() {
		var upload Upload
		if err := rows.Scan(&upload.ID, &upload.Filename, &upload.Username, &upload.Enable); err != nil {
			return nil, err
		}
		uploads = append(uploads, upload)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.New("error during row iteration")
	}

	return uploads, nil
}

// get user data by username from users table
func GetUploadByFileId(file_id uint) (*Upload, error) {
	query := `SELECT id, filename, username, enable, file_key FROM users WHERE id = $1`

	row := DB.QueryRow(context.Background(), query, file_id)

	var data Upload

	err := row.Scan(&data.ID, &data.Filename, &data.Username, &data.Enable, &data.FileKey)
	if err != nil && err.Error() == "no rows in result set" {
		log.Println(err.Error())
		return nil, nil
	} else if err != nil {
		log.Printf("Failed to fetch upload info: %s\n", err.Error())
		return nil, err
	}

	return &data, nil
}

// get user data by username from users table
func UpdateUploadEnableByFileId(file_id uint, enable bool) error {
	query := `UPDATE uploads SET enable = $1 WHERE id = $2`

	_, err := DB.Exec(context.Background(), query, enable, file_id)
	if err != nil {
		log.Printf("Failed to update upload info: %s\n", err.Error())
		return err
	}
	return nil
}
