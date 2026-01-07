package server

import (
	"context"
	"fmt"
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
			c.JSON(400, gin.H{
				"success": false,
				"error":   "Invalid request schema",
			})
			c.Abort()
			util.PrintError(err, "validation err")
			return
		}

		userId, err := bson.ObjectIDFromHex(c.GetString("userId"))
		if err != nil {
			c.JSON(401, gin.H{
				"success": false,
				"error":   "Unauthorized, token missing or invalid",
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
			c.JSON(500, gin.H{
				"success": false,
				"error":   "Internal Server Error",
			})
			c.Abort()
			util.PrintError(err, "class collection insertion err")
			return
		}
		emptyArray := make([]string, 0)

		c.JSON(201, gin.H{
			"success": true,
			"data": gin.H{
				"_id":        res.InsertedID,
				"className":  NewClass.ClassName,
				"teacherId":  userId,
				"studentIds": emptyArray,
			},
		})
	}

}

type AddStudentRequest struct {
	StudentId string `json:"studentId" binding:"required"`
}

func AddStudent(db *mongo.Client) gin.HandlerFunc {

	return func(c *gin.Context) {
		ReqBody := AddStudentRequest{}

		err := c.ShouldBind(&ReqBody)
		if err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"error":   "Invalid request schema",
			})
			c.Abort()
			util.PrintError(err, "validation err")
			return
		}

		collection := db.Database("attendance").Collection("class")

		id, _ := bson.ObjectIDFromHex(c.Param("id"))
		fmt.Println(id)

		filter := bson.M{"_id": id}
		result := &data.Class{}
		err = collection.FindOne(context.Background(), filter).Decode(&result)
		if err != nil {
			c.JSON(404, gin.H{
				"success": false,
				"error":   "Class not found",
			})
			c.Abort()
			util.PrintError(err, "no such class")
			return
		}

		if result.TeacherID.Hex() != c.GetString("userId") {
			c.JSON(403, gin.H{
				"success": false,
				"error":   "Forbidden, not class teacher",
			})
			c.Abort()
			return
		}

		studentId, err := bson.ObjectIDFromHex(ReqBody.StudentId)
		if err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"error":   "Invalid request schema",
			})
			c.Abort()
			util.PrintError(err, "invalid studentId")
		}

		studentFilter := bson.M{
			"_id": studentId,
		}

		var student data.User

		err = db.Database("attendance").Collection("users").FindOne(context.Background(), studentFilter).Decode(&student)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(404, gin.H{
					"success": false,
					"error":   "Student not found",
				})
				util.PrintError(err, "Student finding err")
				c.Abort()
				return
			}
			c.Abort()
			util.PrintError(err, "Student finding err")
			return
		}

		for _, v := range result.StudentIDs {
			if v == studentId {
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data":    result,
				})
				util.PrintError(err, "Student already exists")
				c.Abort()
				return
			}
		}

		update := bson.M{
			"$push": bson.M{
				"student_ids": studentId,
			},
		}

		var updatedClass data.Class
		err = collection.FindOneAndUpdate(context.Background(), filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&updatedClass)
		if err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"error":   "Invalid request schema",
			})
			util.PrintError(err, "Add student db update err")
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    updatedClass,
		})

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
			c.JSON(404, gin.H{
				"success": false,
				"error":   "Class not found",
			})
			c.Abort()
			util.PrintError(err, "db finding err")
			return
		}

		Students := []data.User{}
		if err := cur.All(context.Background(), &Students); err != nil {
			c.JSON(404, gin.H{
				"success": false,
				"error":   "Class not found",
			})
			c.Abort()
			util.PrintError(err, "cursor iteration err")
			return

		}
		//step 3
		//format the response

		type Response struct {
			ID        string                  `json:"_id"`
			ClassName string                  `json:"className"`
			TeacherID string                  `json:"teacherId" `
			Students  []*data.StudentResponse `json:"students" `
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
			"success": true,
			"data":    res,
		})

	}
}
