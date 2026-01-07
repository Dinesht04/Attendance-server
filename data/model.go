package data

import (
	"sync"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID       bson.ObjectID `json:"_id" bson:"_id"`
	Name     string        `json:"name"`
	Email    string        `json:"email"`
	Password string        `json:"password"`
	Role     string        `json:"role"`
}

type Class struct {
	ID         bson.ObjectID   `json:"_id" bson:"_id"`
	ClassName  string          `json:"classname"`
	TeacherID  bson.ObjectID   `json:"teacher_id" bson:"teacher_id"`
	StudentIDs []bson.ObjectID `json:"student_ids" bson:"student_ids"`
}

type AttendanceStatus map[string]string

//validate Role -> teacher | student
//validate Status -> present | absent

type Session struct {
	sync.Mutex
	ClassID          bson.ObjectID
	StartedAt        string
	AttendanceStatus AttendanceStatus
}

type Attendance struct {
	ID        bson.ObjectID `json:"_id" bson:"_id"`
	ClassID   bson.ObjectID `json:"classId"`
	StudentID bson.ObjectID `json:"studentId"`
	Status    string        `json:"status"`
}

type Student struct {
	ID    bson.ObjectID `json:"_id" bson:"_id"`
	Name  string        `json:"name"`
	Email string        `json:"email" `
}

type StudentResponse struct {
	ID    string `json:"_id" bson:"_id"`
	Name  string `json:"name"`
	Email string `json:"email" `
}
