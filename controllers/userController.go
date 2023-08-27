package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"app_backend/db/db"
	helper "app_backend/helpers"
	"app_backend/models"

	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func HashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic("Could not hash password")
	}
	return string(hashedPassword)

}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""
	if err != nil {
		msg = fmt.Sprintf("email of password is incorrect")
		check = false
	}
	return check, msg
}

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validationErr := validator.New().Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
		password := HashPassword(*user.Password)
		user.Password = &password
		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.FirstName, *user.LastName, *user.User_type, *&user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken

		// Assuming you have an edgedb connection named "conn"

		err := db.Pool.Execute(ctx, `
			INSERT User {
				email := <str>$email,
				first_name := <str>$first_name,
				last_name := <str>$last_name,
				user_type := <str>$user_type,
				user_id := <str>$user_id,
				created_at := <datetime>$created_at,
				updated_at := <datetime>$updated_at,
				token := <str>$token,
				refresh_token := <str>$refresh_token
			}`,
		// map[string]interface{}{
		// 	"email":         user.Email,
		// 	"first_name":    user.FirstName,
		// 	"last_name":     user.LastName,
		// 	"user_type":     user.UserType,
		// 	"user_id":       user.UserID,
		// 	"created_at":    user.CreatedAt,
		// 	"updated_at":    user.UpdatedAt,
		// 	"token":         user.Token,
		// 	"refresh_token": user.RefreshToken,

		// },
		)
		if err != nil {
			msg := fmt.Sprintf("user item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return

		}
		defer cancel()
		c.JSON(http.StatusAccepted, &user)
	}
}

// func Login() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
// 		var user models.User
// 		var foundUser models.User

// 		if err := c.BindJSON(&user); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}
// 		err := db.Pool.QuerySingle(ctx, `
// 		SELECT User {
// 			id,
// 			email,
// 			password,
// 			# Other fields you need...
// 		}
// 		FILTER .email = <str>$email`,
// 			&foundUser,
// 			map[string]interface{}{
// 				"email": user.Email,
// 			},
// 		)
// 		defer cancel()
// 		if err != nil {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
// 			return
// 		}

// 		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
// 		defer cancel()
// 		if passwordIsValid != true{
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
// 			return
// 		}
// 		if foundUser.Email == nil{
// 			c.JSON(http.StatusNotFound, gin.H{"error": "No such user exists."})
// 		}
// 		token, refreshToken,_ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.FirstName,*foundUser.LastName, *foundUser.User_type,*&foundUser.User_id)
// 		helper.UpdateAllTokens(token,refreshToken,foundUser.User_id)
// 		err = userCollection.FindOne(ctx, bson.M{"user_id": foundUser.User_id}).Decode(&foundUser)

// 		if err != nil{
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 		c.JSON(http.StatusAccepted, foundUser)

// 	}
// }

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := db.Pool.QuerySingle(ctx, `
			SELECT User {
				id,
				email,
				password,
				first_name,
				last_name,
				user_type,
				user_id
			}
			FILTER .email = <str>$email`,
			&foundUser,
			map[string]interface{}{
				"email": user.Email,
			},
		)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		if !passwordIsValid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			return
		}

		if foundUser.Email == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "No such user exists."})
			return
		}

		// Update tokens in the Edgedb
		err = db.Pool.Execute(ctx, `
			UPDATE User
			FILTER .user_id = <str>$user_id
			SET {
				token := <str>$token,
				refresh_token := <str>$refresh_token
			}`,
		// map[string]interface{}{
		// 	"user_id":        foundUser.UserID,
		// 	"token":          token,
		// 	"refresh_token":  refreshToken,
		// },
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Assuming you've retrieved updated user data from Edgedb after token update
		// Now you can return the updated user data
		c.JSON(http.StatusAccepted, foundUser)
	}
}

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")

		if err := helper.MatchUserTypeToUid(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User

		err := db.Pool.QuerySingle(ctx, `
		SELECT USER { id }
		FILTER .id = <uuid><str>$0
		LIMIT 1`,
			&user,
			c.Param("id"),
		)
		defer cancel()
		if err != nil {
			c.JSON(500, gin.H{"error": "server error"})
		} else {
			c.JSON(200, gin.H{"data": user})
		}
	}
}
