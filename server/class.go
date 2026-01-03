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
			c.JSON(http.StatusOK, gin.H{
				"error": "Invalid Body",
			})
			c.Abort()
			util.PrintError(err, "validation err")
			return
		}

		// if ActiveSession.ClassID.IsZero() == false {
		// 	c.JSON(http.StatusOK, gin.H{
		// 		"error": "Class already in session",
		// 	})
		// 	c.Abort()
		// 	return
		// }

		// ActiveSession.ClassID = bson.NewObjectID()
		// ActiveSession.StartedAt = time.Now().Local().String()
		userId, err := bson.ObjectIDFromHex(c.GetString("userId"))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"error": "internal server error",
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
				"error": "internal server error",
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
				"error": "Invalid Body",
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
				"error": "no such class",
			})
			c.Abort()
			util.PrintError(err, "no such class")
			return
		}

		if result.TeacherID.Hex() != c.GetString("userId") {
			c.JSON(http.StatusOK, gin.H{
				"auth error": "not your class",
			})
			c.Abort()
			fmt.Println(result.TeacherID.String(), " != ", c.GetString("userId"))
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed",
				"err": err})

			return
		}
		c.JSON(http.StatusOK, updatedClass)

	}

}

func GetClass(db *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		//extract class id
		//search query db for teacher id and students id
		//exreact the id from rquest and match
	}
}
