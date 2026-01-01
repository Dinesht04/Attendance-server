package server

import (
	"net/http"

	"github.com/dinesht04/ws-attendance/data"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// func TeacherAuth(params string) gin.HandlerFunc {
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
		class.POST("/")
		class.POST("/:id/add-student")
		class.GET("/:id")
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
