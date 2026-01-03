package server

import (
	"net/http"

	"github.com/dinesht04/ws-attendance/data"
	"github.com/dinesht04/ws-attendance/util"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// the auth flow should be -> req -> extract the token,
// if invalid -> then send bad login response.
// if valid -> add id and role to the context

func Auth() gin.HandlerFunc {
	// <---
	// This is part one
	// --->
	// Example initialization: validate input params
	//this runs when server starts

	return func(c *gin.Context) {
		// this runs during the actual curl
		// <---
		// This is part two
		// --->
		// Example execution per request: inject into context

		header := c.GetHeader("Authorization")
		if header == "" {
			c.JSON(http.StatusAccepted, gin.H{
				"status": "token empty",
			})
			c.Abort()
			return
		}

		token, err := jwt.Parse(header, func(t *jwt.Token) (any, error) {
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil {
			c.JSON(http.StatusAccepted, gin.H{
				"status": "error while parsing, unable to authenticate",
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusAccepted, gin.H{
				"status": "error while claim validation",
			})
			c.Abort()
			return
		}

		role, ok := claims["role"].(string)
		if !ok {
			c.JSON(http.StatusAccepted, gin.H{
				"status": "error while value validation",
			})
			c.Abort()
			return
		}
		userId, ok := claims["userId"].(string)
		if !ok {
			c.JSON(http.StatusAccepted, gin.H{
				"status": "error while value validation",
			})
			c.Abort()
			return
		}

		if role == "student" || role == "teacher" {
			c.Set("role", role)
			c.Set("userId", userId)
			c.Next()
		} else {
			c.JSON(http.StatusAccepted, gin.H{
				"status": "auth err who r u br",
			})
			c.Abort()
			return
		}

	}
}

func TeacherRoleAuth() gin.HandlerFunc {

	return func(c *gin.Context) {
		if c.GetString("role") == "teacher" {
			c.Next()
		} else {
			c.Abort()
			c.JSON(http.StatusOK, gin.H{
				"auth error": "you are not a teacher",
			})
			return
		}
	}

}

func StudentRoleAuth(c *gin.Context) {
	if c.GetString("role") == "student" {
		c.Next()
	} else {
		c.Abort()
		c.JSON(http.StatusOK, gin.H{
			"auth error": "you are not a student? why are you as a teacher accessing it maaan",
		})
		return
	}
}

func ClassBasedAuth(db *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		classId, err := bson.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"eror": "internal server error",
			})
			c.Abort()
			util.PrintError(err, "object id err")
			return
		}

		Class := data.Class{}

		filter := bson.M{"_id": classId}

		err = db.Database("attendance").Collection("class").FindOne(c, filter).Decode(&Class)

		if err != nil {

			c.JSON(http.StatusOK, gin.H{
				"eror": "internal server error",
			})
			c.Abort()
			util.PrintError(err, "db finding err")
			return
		}

		c.Set("studentIds", Class.StudentIDs)
		c.Set("classId", Class.ID.Hex())
		c.Set("teacherId", Class.TeacherID.Hex())
		c.Set("className", Class.ClassName)

		userId, err := bson.ObjectIDFromHex(c.GetString("userId"))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"eror": "internal server error",
			})
			c.Abort()
			util.PrintError(err, "object id err")
			return
		}

		if c.GetString("role") == "teacher" {

			if userId.Hex() != Class.TeacherID.Hex() {
				c.JSON(http.StatusOK, gin.H{
					"auth err": "not your class bru",
				})
				c.Abort()
				return
			}
			c.Next()
		}

		if c.GetString("role") == "student" {

			for _, v := range Class.StudentIDs {
				if userId == v {
					c.Next()
				}
			}
			c.JSON(http.StatusOK, gin.H{
				"auth err": "not your class bru",
			})
			c.Abort()
			return

		}
	}
}

// func StudentAuth(params string) gin.HandlerFunc {
// 	// <---
// 	// This is part one
// 	// --->
// 	// Example initialization: validate input params
// 	if err := check(params); err != nil {
// 		panic(err)
// 	}
// 	//this runs when server starts

// 	return func(c *gin.Context) {
// 		// this runs during the actual curl
// 		// <---
// 		// This is part two
// 		// --->
// 		// Example execution per request: inject into context
// 		c.Set("TestVar", params)
// 		c.Next()
// 	}
// }

var ActiveSession data.Session

func StartServer(db *mongo.Client) {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	{
		auth := r.Group("/auth")
		auth.POST("/signup", HandleSignup(db))
		auth.POST("/login", HandleLogin(db))
		auth.POST("/me", HandleMe(db))
	}

	{
		class := r.Group("/class")
		class.POST("/", Auth(), TeacherRoleAuth(), CreateClass(db))
		class.POST("/:id/add-student", Auth(), TeacherRoleAuth(), AddStudent(db))
		class.GET("/:id/", Auth(), ClassBasedAuth(db), GetClass(db))
		class.GET("/:id/my-attendance")
	}

	{
		students := r.Group("/students")
		students.GET("/")
	}

	{
		attendance := r.Group("/attendance")
		attendance.POST("/start")
	}

	r.Run()
}
