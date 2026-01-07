package server

import (
	"context"
	"net/http"

	"github.com/dinesht04/ws-attendance/data"
	"github.com/dinesht04/ws-attendance/util"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type CreateClassRequest struct {
	ClassName string `json:"className" binding:"required"`
}

func CreateClass(db *mongo.Client) gin.HandlerFunc {

	return func(c *gin.Context) {
		ReqBody := CreateClassRequest{}

		err := c.ShouldBind(&ReqBody)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "Invalid request schema",
			})
			c.Abort()
			util.PrintError(err, "validation err")
			return
		}

		userId, err := bson.ObjectIDFromHex(c.GetString("userId"))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "Internal Server Error",
			})
			c.Abort()
			util.PrintError(err, "bson from hex err")
			return

		}

		NewClass := data.Class{
			ID:         bson.NewObjectID(),
			ClassName:  ReqBody.ClassName,
			TeacherID:  userId,
			StudentIDs: []bson.ObjectID{},
		}

		res, err := db.Database("attendance").Collection("class").InsertOne(context.Background(), &NewClass)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "Internal Server Error",
			})
			c.Abort()
			util.PrintError(err, "class collection insertion err")
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": "true",
			"data": gin.H{
				"_id":        res.InsertedID,
				"className":  NewClass.ClassName,
				"teacherId":  userId,
				"studentIds": gin.H{},
			},
		})
	}

}

type AddStudentRequest struct {
	StudentId string `json:"studentId"`
}

func AddStudent(db *mongo.Client) gin.HandlerFunc {

	return func(c *gin.Context) {
		ReqBody := AddStudentRequest{}

		err := c.ShouldBind(&ReqBody)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "Invalid request schema",
			})
			c.Abort()
			util.PrintError(err, "validation err")
			return
		}

		collection := db.Database("attendance").Collection("class")

		id, _ := bson.ObjectIDFromHex(c.Param("id"))
		filter := bson.M{"_id": id}
		result := &data.Class{}
		err = collection.FindOne(context.Background(), filter).Decode(&result)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "Class not found",
			})
			c.Abort()
			util.PrintError(err, "no such class")
			return
		}

		if result.TeacherID.Hex() != c.GetString("userId") {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "Forbidden, not class teacher",
			})
			c.Abort()
			return
		}

		studentId, _ := bson.ObjectIDFromHex(ReqBody.StudentId)

		update := bson.M{
			"$push": bson.M{
				"student_ids": studentId,
			},
		}

		var updatedClass bson.M
		err = collection.FindOneAndUpdate(context.Background(), filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&updatedClass)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Internal server error",
			})
			return
		}
		c.JSON(http.StatusOK, updatedClass)

	}

}

func GetClass(db *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		studentsIds, _ := c.Get("studentIds")

		filter := bson.M{
			"_id": bson.M{
				"$in": studentsIds,
			},
		}

		cur, err := db.Database("attendance").Collection("users").Find(context.Background(), filter)

		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "Class not found",
			})
			c.Abort()
			util.PrintError(err, "db finding err")
			return
		}

		Students := []data.User{}
		if err := cur.All(context.Background(), &Students); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "Internal Server Error",
			})
			c.Abort()
			util.PrintError(err, "cursor iteration err")
			return

		}
		//step 3
		//format the response

		type Response struct {
			ID        string                  `json:"_id"`
			ClassName string                  `json:"classname"`
			TeacherID string                  `json:"teacher_id" `
			Students  []*data.StudentResponse `json:"studens" `
		}

		res := &Response{
			ID:        c.GetString("classId"),
			ClassName: c.GetString("className"),
			TeacherID: c.GetString("teacherId"),
		}

		for _, v := range Students {
			res.Students = append(res.Students, &data.StudentResponse{
				ID:    v.ID.Hex(),
				Name:  v.Name,
				Email: v.Email,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"success": "true",
			"data":    res,
		})

	}
}
