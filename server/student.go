package server

import (
	"net/http"

	"github.com/dinesht04/ws-attendance/data"
	"github.com/dinesht04/ws-attendance/util"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func getStudents(db *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		filter := bson.M{
			"role": "student",
		}

		cur, err := db.Database("attendance").Collection("users").Find(c, filter)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "Students not found",
			})
			util.InternalServerError(c, err, "collection finding err")
			return
		}
		Students := []data.Student{}
		if err := cur.All(c, &Students); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "Internal Server Error",
			})
			util.InternalServerError(c, err, "cur iteration err")
			return
		}
		type Response struct {
			Success bool                    `json:"success"`
			Data    []*data.StudentResponse `json:"data"`
		}
		res := &Response{
			Success: true,
		}

		for _, v := range Students {
			id := v.ID.Hex()
			res.Data = append(res.Data, &data.StudentResponse{
				ID:    id,
				Name:  v.Name,
				Email: v.Email,
			})
		}
		c.JSON(http.StatusOK, res)
	}
}
