package helper

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/mikekbnv/To-Do-List/database"
	"github.com/mikekbnv/To-Do-List/internal/model"

	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	jwt.StandardClaims
}

var usersCollection *mongo.Collection = database.OpenCollection("users")
var SECRET_KEY string = os.Getenv("SECRET_KEY")

// GenerateAllTokens generates both teh detailed token and refresh token
func GenerateAllTokens() (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Minute * time.Duration(5)).Unix(),
		},
	}

	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(1)).Unix(),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return
	}

	return token, refreshToken, err
}

//ValidateToken validates the jwt token
func ValidateToken(signedToken, refresh_token string) (user model.User, msg string) {
	_, token_msg := parseclaims(signedToken)
	_, refresh_token_msg := parseclaims(refresh_token)

	if token_msg == "" {
		user, err := get_user_by_token(signedToken, "token")
		if err == nil {
			return user, ""
		} else {
			return user, "the token did not find in DB"
		}

	} else {
		if refresh_token_msg == "" {
			user, err := get_user_by_token(refresh_token, "refresh_token")
			if err == nil {
				new_token, new_refresh_token, _ := GenerateAllTokens()
				UpdateAllTokens(new_token, new_refresh_token, user.User_id)
				user.Token = &new_token
				user.Refresh_token = &new_refresh_token
				return user, ""
			} else {
				return user, "the refresh token did not find in DB" + refresh_token_msg
			}
		} else {
			return user, "token and refresh_token cannot be parsed: " + token_msg + refresh_token_msg
		}
	}
}

//UpdateAllTokens renews the user tokens when they login
func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	var updateObj primitive.D

	updateObj = append(updateObj, bson.E{"token", signedToken})
	updateObj = append(updateObj, bson.E{"refresh_token", signedRefreshToken})

	Updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{"updated_at", Updated_at})

	upsert := true
	filter := bson.M{"user_id": userId}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := usersCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{"$set", updateObj},
		},
		&opt,
	)
	defer cancel()

	if err != nil {
		log.Panic("Update tokens error:", err)
		return
	}
}

func parseclaims(token string) (claims jwt.StandardClaims, msg string) {
	t, err := jwt.ParseWithClaims(
		token,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)

	if err != nil {
		msg = err.Error()
		return
	}

	cl, ok := t.Claims.(*SignedDetails)
	if !ok {
		return cl.StandardClaims, "the token is invalid"
	}
	return cl.StandardClaims, ""
}

func get_user_by_token(token, name string) (model.User, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var user model.User

	find := usersCollection.FindOne(ctx, bson.M{name: token}).Decode(&user)
	return user, find
}
