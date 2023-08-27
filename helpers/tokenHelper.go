package helper

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"app_backend/db/db"
	jwt "github.com/dgrijalva/jwt-go"
)

type SignedDetails struct {
	Email     string
	FirstName string
	LastName  string
	Uid       string
	User_type string
	jwt.StandardClaims
}

var SECRET_KEY string = os.Getenv("SECRET_KEY")

func GenerateAllTokens(email string, first_name string, last_name string, user_type string, uid string) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:     email,
		FirstName: first_name,
		LastName:  last_name,
		Uid:       uid,
		User_type: user_type,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Minute * time.Duration(168)).Unix(),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	refresh_token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return
	}
	return token, refresh_token, err
}

// func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string){
// 	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

// 	var updateObj primitive.D

// 	updateObj = append(updateObj, bson.E{"token", signedToken})
// 	updateObj = append(updateObj, bson.E{"refresh_token", signedRefreshToken})

// 	Updated_at, _ := time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
// 	updateObj = append(updateObj, bson.E{"updated_at", Updated_at})

// 	upsert := true
// 	filter := bson.M{"user_id":userId}
// 	opt := options.UpdateOptions{
// 		Upsert: &upsert,
// 	}
// 	_, err := collection.UpdateOne(ctx, filter, bson.D{"$set",updateObj},opt)
// 	defer cancel()

// 	if err != nil{
// 		log.Println("Error while updating the tokens")
// 		return
// 	}
// 	return
// }

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(
		signedToken, &SignedDetails{},
		func(t *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)
	if err != nil {
		msg = err.Error()
		return
	}
	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = fmt.Sprintf("the token is invalid")
		msg = err.Error()
		return
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = fmt.Sprintf("token is expired")
		msg = err.Error()
		return
	}
	return claims, msg
}

func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	updateQuery := `
		UPDATE User
		FILTER .user_id = <str>$user_id
		SET {
			token := <str>$token,
			refresh_token := <str>$refresh_token,
			updated_at := <datetime>$updated_at
		}
	`

	err := db.Pool.Execute(ctx, updateQuery) // map[string]interface{}{
	// 	"user_id":      userId,
	// 	"token":        signedToken,
	// 	"refresh_token": signedRefreshToken,
	// 	"updated_at":   time.Now(),
	// },

	if err != nil {
		log.Println("Error while updating the tokens:", err)
		return
	}
}
