package server

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/dinesht04/ws-attendance/data"
	"github.com/dinesht04/ws-attendance/util"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// the auth flow should be -> req -> extract the token,
// if invalid -> then send bad login response.
// if valid -> add id and role to the context

type ConnectionStruct struct {
	authenticated bool
	ws            *websocket.Conn
}

type OpenSocketsStruct struct {
	sync.Mutex
	list map[string]*ConnectionStruct
}

func (p *OpenSocketsStruct) Add(id string, ws *websocket.Conn) {
	p.Lock()
	defer p.Unlock()

	p.list = map[string]*ConnectionStruct{}
	p.list[id] = &ConnectionStruct{
		authenticated: true,
		ws:            ws,
	}
}

func (p *OpenSocketsStruct) SendMessage() {
	fmt.Println(p.list)
}

var SocketList OpenSocketsStruct

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
			c.JSON(401, gin.H{
				"success": false,
				"error":   "Unauthorized, token missing or invalid",
			})
			c.Abort()
			return
		}

		token, err := jwt.Parse(header, func(t *jwt.Token) (any, error) {
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil {
			c.JSON(401, gin.H{
				"success": false,
				"error":   "Unauthorized, token missing or invalid",
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(401, gin.H{
				"success": false,
				"error":   "Unauthorized, token missing or invalid",
			})
			c.Abort()
			return
		}

		role, ok := claims["role"].(string)
		if !ok {
			c.JSON(401, gin.H{
				"success": false,
				"error":   "Unauthorized, token missing or invalid",
			})
			c.Abort()
			return
		}
		userId, ok := claims["userId"].(string)
		if !ok {
			c.JSON(401, gin.H{
				"success": false,
				"error":   "Unauthorized, token missing or invalid",
			})
			c.Abort()
			return
		}

		if role == "student" || role == "teacher" {
			c.Set("role", role)
			c.Set("userId", userId)
			c.Next()
		} else {
			c.JSON(401, gin.H{
				"success": false,
				"error":   "Unauthorized, token missing or invalid",
			})
			c.Abort()
			return
		}

	}
}

func QueryParamsAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		jwToken, exists := c.GetQuery("token")
		if !exists {
			util.AuthError(c, fmt.Errorf("JWT token not present"))
			return
		}

		token, err := jwt.Parse(jwToken, func(t *jwt.Token) (any, error) {
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil {
			c.JSON(401, gin.H{
				"success": false,
				"error":   "Unauthorized, token missing or invalid",
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(401, gin.H{
				"success": false,
				"error":   "Unauthorized, token missing or invalid",
			})
			c.Abort()
			return
		}
		role, ok := claims["role"].(string)
		if !ok {
			c.JSON(401, gin.H{
				"success": false,
				"error":   "Unauthorized, token missing or invalid",
			})
			c.Abort()
			return
		}
		userId, ok := claims["userId"].(string)
		if !ok {
			c.JSON(401, gin.H{
				"success": false,
				"error":   "Unauthorized, token missing or invalid",
			})
			c.Abort()
			return
		}

		if role == "student" || role == "teacher" {
			c.Set("role", role)
			c.Set("userId", userId)
			c.Next()
		} else {
			c.JSON(401, gin.H{
				"success": false,
				"error":   "Unauthorized, token missing or invalid",
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
			c.JSON(403, gin.H{
				"success": false,
				"error":   "Forbidden, teacher access required",
			})
			c.Abort()
			return
		}
	}

}

func StudentRoleAuth() gin.HandlerFunc {

	return func(c *gin.Context) {
		if c.GetString("role") == "student" {
			c.Next()
		} else {
			c.Abort()
			c.JSON(403, gin.H{
				"success": false,
				"error":   "Forbidden, teacher access required",
			})
			c.Abort()
			return
		}
	}

}

func ClassParamBasedAuth(db *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		classId, err := bson.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(401, gin.H{
				"success": false,
				"error":   "Unauthorized, token missing or invalid",
			})
			c.Abort()
			util.PrintError(err, "object id err")
			return
		}

		Class := data.Class{}

		filter := bson.M{"_id": classId}

		err = db.Database("attendance").Collection("class").FindOne(c, filter).Decode(&Class)

		if err != nil {

			c.JSON(404, gin.H{
				"success": false,
				"error":   "Class not found",
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
			c.JSON(401, gin.H{
				"success": false,
				"error":   "Unauthorized, token missing or invalid",
			})
			c.Abort()
			util.PrintError(err, "object id err")
			return
		}

		if c.GetString("role") == "teacher" {

			if userId.Hex() != Class.TeacherID.Hex() {
				c.JSON(403, gin.H{
					"success": false,
					"error":   "Forbidden, not class teacher",
				})
				c.Abort()
				return
			}
			c.Next()
		} else if c.GetString("role") == "student" {
			verified := false
			for _, v := range Class.StudentIDs {
				if userId == v {
					verified = true
					c.Next()
				}
			}
			if verified == false {
				c.JSON(403, gin.H{
					"success": false,
					"error":   "Forbidden, not class teacher",
				})
				c.Abort()
				return
			}

		}
	}
}

func ClassBodyBasedAuth(db *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		type StartReq struct {
			ClassID string `json:"classId" binding:"required"`
		}

		req := StartReq{}
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"error":   "Invalid request schema",
			})
			util.PrintError(err, "req bind err")
			c.Abort()
			return
		}

		classId, err := bson.ObjectIDFromHex(req.ClassID)
		if err != nil {
			c.JSON(401, gin.H{
				"success": false,
				"error":   "Unauthorized, token missing or invalid",
			})
			c.Abort()
			util.PrintError(err, "object id err")
			return
		}

		Class := data.Class{}

		filter := bson.M{"_id": classId}

		err = db.Database("attendance").Collection("class").FindOne(c, filter).Decode(&Class)

		if err != nil {

			c.JSON(404, gin.H{
				"success": false,
				"error":   "Class not found",
			})
			c.Abort()
			util.PrintError(err, "db finding err")
			return
		}

		c.Set("studentIds", Class.StudentIDs)
		c.Set("classId", Class.ID)
		c.Set("teacherId", Class.TeacherID.Hex())
		c.Set("className", Class.ClassName)

		userId, err := bson.ObjectIDFromHex(c.GetString("userId"))
		if err != nil {
			c.JSON(401, gin.H{
				"success": false,
				"error":   "Unauthorized, token missing or invalid",
			})
			c.Abort()
			util.PrintError(err, "object id err")
			return
		}

		if c.GetString("role") == "teacher" {

			if userId.Hex() != Class.TeacherID.Hex() {
				c.JSON(403, gin.H{
					"success": false,
					"error":   "Forbidden, not class teacher",
				})
				c.Abort()
				return
			}
			c.Next()
		} else if c.GetString("role") == "student" {
			verified := false
			for _, v := range Class.StudentIDs {
				if userId == v {
					verified = true
					c.Next()
				}
			}
			if verified == false {
				c.JSON(403, gin.H{
					"success": false,
					"error":   "Forbidden, not class teacher",
				})
				c.Abort()
				return
			}

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

	ActiveSession.AttendanceStatus = make(data.AttendanceStatus)

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
		class := r.Group("/class", Auth())
		class.POST("/", TeacherRoleAuth(), CreateClass(db))
		class.POST("/:id/add-student", TeacherRoleAuth(), AddStudent(db))
		class.GET("/:id/", ClassParamBasedAuth(db), GetClass(db))
		class.GET("/:id/my-attendance", StudentRoleAuth(), ClassParamBasedAuth(db), getMyAttendance(db))
	}

	{
		students := r.Group("/students", Auth())
		students.GET("/", TeacherRoleAuth(), getStudents(db))
	}

	{
		attendance := r.Group("/attendance", Auth())
		attendance.POST("/start", TeacherRoleAuth(), ClassBodyBasedAuth(db), startAttendance(db))
	}

	hub := &Hub{
		Clients:    make(map[*Client]bool),
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		db:         db,
	}

	go hub.Run()

	{
		ws := r.Group("/ws")
		ws.GET("/", QueryParamsAuth(), handleWebsocket(db, hub))
	}

	r.Run()
}
