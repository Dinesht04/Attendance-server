package server

import (
	"context"
	"net/http"

	"github.com/dinesht04/ws-attendance/data"
	"github.com/dinesht04/ws-attendance/util"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

const secret = "golang"

type SignUpRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,gte=6"`
	Role     string `json:"role" binding:"required,oneof=student teacher"`
}

func HandleSignup(db *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqBody := SignUpRequest{}
		if err := c.ShouldBind(&reqBody); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "Invalid request schema",
			})
			util.PrintError(err, "Validaton and Binding err")
			return
		}

		hashedPass, err := bcrypt.GenerateFromPassword([]byte(reqBody.Password), 5)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "Internal server error",
			})
			util.PrintError(err, "Password Hashing error")
			return
		}

		// 	ID       bson.ObjectID `json:"_id" bson:"_id"`
		// Name     string        `json:"name"`
		// Email    string        `json:"email"`
		// Password string        `json:"password"`
		// Role     string        `json:"role"`

		collection := db.Database("attendance").Collection("users")

		filter := bson.M{"email": reqBody.Email}
		User := data.User{}

		err = collection.FindOne(context.Background(), filter).Decode(&User)
		if err == nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "Email already exists",
			})
			return
		}

		res, err := collection.InsertOne(context.Background(), bson.M{
			"name":     reqBody.Name,
			"email":    reqBody.Email,
			"password": hashedPass,
			"role":     reqBody.Role,
		})
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "Internal Server Error",
			})
			util.PrintError(err, "DB insertion err")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"data": gin.H{
				"_id":   res.InsertedID,
				"name":  reqBody.Name,
				"email": reqBody.Email,
				"role":  reqBody.Role,
			},
		})
	}

}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,gte=6"`
}

func HandleLogin(db *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqBody := LoginRequest{}
		if err := c.ShouldBind(&reqBody); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "Invalid request schema",
			})
			util.PrintError(err, "Validaton and Binding err")
			return
		}

		filter := bson.M{"email": reqBody.Email}
		User := data.User{}

		collection := db.Database("attendance").Collection("users")

		err := collection.FindOne(context.Background(), filter).Decode(&User)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "User not found",
			})
			util.PrintError(err, "Searching for user err")
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(User.Password), []byte(reqBody.Password))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "Invalid email or password",
			})
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"userId": User.ID,
			"role":   User.Role,
		})

		// Sign and get the complete encoded token as a string using the secret
		tokenString, err := token.SignedString([]byte(secret))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "Internal Server Error",
			})
			util.PrintError(err, "SIgning jwt err")
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"token": tokenString,
			},
		})
	}

}

type MyClaims struct {
	UserId string `json:"userId"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func HandleMe(db *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Unauthorized, token missing or invalid",
			})
			return
		}

		//parse jwt here
		token, err := jwt.Parse(authHeader, func(token *jwt.Token) (any, error) {
			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "Unauthorized, token missing or invalid",
			})
			util.PrintError(err, "validating jwt error")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "Unauthorized, token missing or invalid",
			})
			util.PrintError(err, "extracting claims error")
			return
		}

		userId, ok := claims["userId"].(string)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Unauthorized, token missing or invalid",
			})
			util.PrintError(err, "Token is not string")
			return
		}

		objId, _ := bson.ObjectIDFromHex(userId)

		filter := bson.M{"_id": objId}
		User := data.User{}

		collection := db.Database("attendance").Collection("users")

		err = collection.FindOne(context.Background(), filter).Decode(&User)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "User not found",
			})
			util.PrintError(err, "Searching for user err")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"_id":   User.ID,
				"name":  User.Name,
				"email": User.Email,
				"role":  User.Role,
			},
		})
	}

}
