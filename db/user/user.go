package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/edgedb/edgedb-go"
)

type User struct {
	ID            int       `json:"id"`
	FirstName     *string   `json:"first_name" validate: required, min=2, max=100`
	LastName      *string   `json:"last_name" validate: required, min=2, max=100`
	Email         *string   `json:"email" validate:"email, required"`
	Password      *string   `json:"password" validate:"required,min=6` // Note: Password should be hashed and salted in production
	Token         *string   `json:"token"`
	Refresh_token *string   `json:"refresh_token"`
	Created_at    time.Time `json:"created_at"`
	Updated_at    time.Time `json:"updated_at"`
	User_id       string    `json:"user_id"`
}

func main() {
	opts := edgedb.Options{Concurrency: 4}
	ctx := context.Background()
	db, err := edgedb.CreateClientDSN(ctx, "edgedb://edgedb@localhost/test", opts)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// create a user object type.
	err = db.Execute(ctx, `
		CREATE TYPE User {
			CREATE REQUIRED PROPERTY name -> str;
			CREATE PROPERTY dob -> datetime;
		}
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Insert a new user.
	var inserted struct{ id edgedb.UUID }
	err = db.QuerySingle(ctx, `
		INSERT User {
			name := <str>$0,
			dob := <datetime>$1
		}
	`, &inserted, "Bob", time.Date(1984, 3, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		log.Fatal(err)
	}

	// Select users.
	var users []User
	args := map[string]interface{}{"name": "Bob"}
	query := "SELECT User {name, dob} FILTER .name = <str>$name"
	err = db.Query(ctx, query, &users, args)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(users)
}
