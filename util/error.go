package util

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func PrintError(err error, message ...string) {
	if len(message) == 0 {
		fmt.Println(fmt.Errorf("Encountered Error: %w", err))
	} else {
		fmt.Println(message[0], " : ", err)
	}
}

func InternalServerError(c *gin.Context, err error, message ...string) {
	c.JSON(http.StatusOK, gin.H{
		"eror": "internal server error",
	})
	c.Abort()
	PrintError(err, message...)
}

func AuthError(c *gin.Context, err error, message ...string) {
	c.JSON(401, gin.H{
		"success": false,
		"error":   "Unauthorized, token missing or invalid",
	})
	c.Abort()
	PrintError(err, message...)
}

// func wsError(msg ...string) *server.WsReq {
// 	err := msg[0]

// 	return &server.WsReq{
// 		Event: "Event error",
// 		Error: &err,
// 	}
// }
