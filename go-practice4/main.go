package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type User struct {
	ID      int     `db:"id"`
	Name    string  `db:"name"`
	Email   string  `db:"email"`
	Balance float64 `db:"balance"`
}

func connect() (*sqlx.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "user=user password=password dbname=mydatabase host=localhost port=5430 sslmode=disable"
	}
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	return db, nil
}

func InsertUser(db *sqlx.DB, user User) error {
	const q = `INSERT INTO users (name, email, balance) VALUES (:name, :email, :balance)`
	_, err := db.NamedExec(q, user)
	return err
}

func GetAllUsers(db *sqlx.DB) ([]User, error) {
	var users []User
	err := db.Select(&users, `SELECT id, name, email, balance FROM users ORDER BY id`)
	return users, err
}

func GetUserByID(db *sqlx.DB, id int) (User, error) {
	var u User
	err := db.Get(&u, `SELECT id, name, email, balance FROM users WHERE id=$1`, id)
	return u, err
}

func TransferBalance(db *sqlx.DB, fromID, toID int, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be > 0")
	}
	if fromID == toID {
		return errors.New("cannot transfer to the same user")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	var sender User
	if err := tx.GetContext(ctx, &sender, `
		SELECT id, name, email, balance
		FROM users
		WHERE id = $1
		FOR UPDATE
	`, fromID); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("sender not found: %w", err)
	}

	var receiver User
	if err := tx.GetContext(ctx, &receiver, `
		SELECT id, name, email, balance
		FROM users
		WHERE id = $1
		FOR UPDATE
	`, toID); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("receiver not found: %w", err)
	}

	if sender.Balance < amount {
		_ = tx.Rollback()
		return errors.New("insufficient funds")
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE users SET balance = balance - $1 WHERE id = $2
	`, amount, sender.ID); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("debit sender: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE users SET balance = balance + $1 WHERE id = $2
	`, amount, receiver.ID); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("credit receiver: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func main() {
	db, err := connect()
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer db.Close()

	_ = InsertUser(db, User{Name: "Alice", Email: "alice@example.com", Balance: 100})
	_ = InsertUser(db, User{Name: "Bob", Email: "bob@example.com", Balance: 50})

	users, err := GetAllUsers(db)
	if err != nil {
		log.Println("GetAllUsers error:", err)
	} else {
		log.Println("Users:", users)
	}

	if err := TransferBalance(db, 1, 2, 20); err != nil {
		log.Println("TransferBalance error:", err)
	} else {
		log.Println("TransferBalance: success")
	}
}
