package models

import "time"

type User struct {
	ID            int       `edgedb:"id" json:"id"`
	FirstName     *string   `edgedb:"first_name" json:"first_name" validate: required, min=2, max=100`
	LastName      *string   `edgedb:"last_name" json:"last_name" validate: required, min=2, max=100`
	Email         *string   `edgedb:"email" json:"email" validate:"email, required"`
	Password      *string   `edgedb:"password" json:"password" validate:"required,min=6` // Note: Password should be hashed and salted in production
	Token         *string   `edgedb:"token" json:"token"`
	User_type     *string   `edgedb:"user_type" json:"user_type"`
	Refresh_token *string   `edgedb:"refresh_token" json:"refresh_token"`
	Created_at    time.Time `edgedb:"created_at" json:"created_at"`
	Updated_at    time.Time `edgedb:"updated_at" json:"updated_at"`
	User_id       string    `edgedb:"user_id" json:"user_id"`
}
