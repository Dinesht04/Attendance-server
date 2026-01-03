package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dinesht04/ws-attendance/data"
	"github.com/dinesht04/ws-attendance/util"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type AttendanceRecord struct {
	ClassID string  `json:"classId"`
	Status  *string `json:"status"`
}

type Response struct {
	Success bool              `json:"success"`
	Data    *AttendanceRecord `json:"data"`
}

func getMyAttendance(db *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := bson.ObjectIDFromHex(c.GetString("userId"))
		if err != nil {
			util.InternalServerError(c, err, "object id from hex err")
			return
		}

		filter := bson.M{
			"_id": id,
		}

		attendance := data.Attendance{}

		err = db.Database("attendance").Collection("attedances").FindOne(c, filter).Decode(&attendance)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusOK, &Response{
					Success: true,
					Data: &AttendanceRecord{
						ClassID: c.GetString("classId"),
						Status:  nil,
					},
				})
				c.Abort()
				return
			}

			util.InternalServerError(c, err, "db search err")
			return
		}

		c.JSON(http.StatusOK, &Response{
			Success: true,
			Data: &AttendanceRecord{
				ClassID: c.GetString("classId"),
				Status:  &attendance.Status,
			},
		})

	}
}

func startAttendance(db *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		class, exists := c.Get("classId")
		if !exists {
			util.InternalServerError(c, fmt.Errorf("class id doesnt exist"))
		}
		classId := class.(bson.ObjectID)

		ActiveSession.ClassID = classId
		ActiveSession.StartedAt = time.Now().UTC().String()

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"classId":   ActiveSession.ClassID,
				"startedAt": ActiveSession.StartedAt,
			},
		})

	}
}
