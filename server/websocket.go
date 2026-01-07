package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dinesht04/ws-attendance/data"
	"github.com/dinesht04/ws-attendance/util"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var upgrader = websocket.Upgrader{}

type Data struct {
	StudentID *string `json:"studentID"`
	Status    *string `json:"status"`
}

type WsReq struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type AttendanceData struct {
	StudentID string `json:"studentID"`
	Status    string `json:"status"`
}

type WsAttendanceMarkReq struct {
	Event string         `json:"event"`
	Data  AttendanceData `json:"data"`
}

type WsTodaySummaryData struct {
	Present int `json:"present"`
	Absent  int `json:"absent"`
	Total   int `json:"total"`
}

type WsTodaySummary struct {
	Event string             `json:"event"`
	Data  WsTodaySummaryData `json:"data"`
}

type WsMyAttendanceData struct {
	Status string `json:"status"`
}

type WsMyAttendance struct {
	Event string             `json:"event"`
	Data  WsMyAttendanceData `json:"data"`
}

type WsDoneData struct {
	Message string `json:"message"`
	Present int    `json:"present"`
	Absent  int    `json:"absent"`
	Total   int    `json:"total"`
}

type WsDone struct {
	Event string     `json:"event"`
	Data  WsDoneData `json:"data"`
}

type WsErrorData struct {
	Message string `json:"message"`
}

type wsError struct {
	Event string      `json:"event"`
	Data  WsErrorData `json:"data"`
}

type WsEvent interface {
	EventName() string
}

func (w WsAttendanceMarkReq) EventName() string { return w.Event }
func (w WsTodaySummary) EventName() string      { return w.Event }
func (w WsMyAttendance) EventName() string      { return w.Event }
func (w WsDone) EventName() string              { return w.Event }
func (w wsError) EventName() string             { return w.Event }
func (w WsReq) EventName() string               { return w.Event }

func handleWebsocket(db *mongo.Client, h *Hub) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		w, r := ctx.Writer, ctx.Request
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			util.InternalServerError(ctx, err, "upgrade err")
			return
		}

		client := &Client{
			id:   ctx.GetString("userId"),
			hub:  h,
			conn: c,
			send: make(chan WsEvent, 256),
			role: ctx.GetString("role"),
		}

		client.hub.register <- client

		go client.writePump()
		go client.readPump(db)

	}
}

func (c *Client) readPump(db *mongo.Client) {
	defer func() {
		c.conn.Close()
		c.hub.unregister <- c
	}()

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			fmt.Println(err, "reading msg error")
		}

		valid := json.Valid(msg)
		fmt.Println(string(msg))
		if !valid {
			fmt.Println("json is valid: ", valid)
			errMsg := wsError{
				Event: "ERROR",
				Data: WsErrorData{
					Message: "Invalid JSON",
				},
			}
			c.send <- errMsg
			continue
		}
		req := WsReq{}

		err = json.Unmarshal(msg, &req)
		if err != nil {
			errMsg := wsError{
				Event: "ERROR",
				Data: WsErrorData{
					Message: "Invalid JSON",
				},
			}
			c.send <- errMsg

			continue
		}

		//this is where we handle everything

		switch req.Event {
		case "ATTENDANCE_MARKED":
			if c.role != "teacher" {

				errMsg := wsError{
					Event: "ERROR",
					Data: WsErrorData{
						Message: "Forbidden, teacher event only",
					},
				}

				c.send <- errMsg
			} else if ActiveSession.StartedAt == "" {

				errMsg := &wsError{
					Event: "ERROR",
					Data: WsErrorData{
						Message: "No active attendance session",
					},
				}

				c.send <- errMsg
			} else {

				jsonData, _ := json.Marshal(req)
				var attendance WsAttendanceMarkReq
				if err := json.Unmarshal(jsonData, &attendance); err != nil {
					c.send <- wsError{Event: "ERROR", Data: WsErrorData{Message: "Invalid format"}}
				}

				ActiveSession.Lock()
				ActiveSession.AttendanceStatus[attendance.Data.StudentID] = attendance.Data.Status
				ActiveSession.Unlock()

				message := &Message{
					Type:     "ATTENDANCE_MARKED",
					ClientID: c.id,
					Text:     attendance,
				}

				c.hub.broadcast <- message
			}

		case "TODAY_SUMMARY":
			if c.role != "teacher" {
				errMsg := &wsError{
					Event: "Auth eroor",
					Data: WsErrorData{
						Message: "Forbidden, teacher event only",
					},
				}

				c.send <- errMsg
			} else if ActiveSession.StartedAt == "" {

				errMsg := &wsError{
					Event: "Event error",
					Data: WsErrorData{
						Message: "No active attendance sessions",
					},
				}

				c.send <- *errMsg
			} else {
				absent := 0
				present := 0
				total := 0
				for _, v := range ActiveSession.AttendanceStatus {
					switch v {
					case "present":
						present++
					case "absent":
						absent++
					default:
						continue
					}
				}
				total = present + absent

				jsonData, _ := json.Marshal(req)
				var today_summary WsTodaySummary
				err := json.Unmarshal(jsonData, &today_summary)
				if err != nil {
					c.send <- wsError{Event: "ERROR", Data: WsErrorData{Message: "Invalid format"}}
				}

				wsMsg := Message{
					ClientID: c.id,
					Text: WsTodaySummary{
						Event: "TODAY_SUMMARY",
						Data: WsTodaySummaryData{
							Present: present,
							Absent:  absent,
							Total:   total,
						},
					},
				}
				c.hub.broadcast <- &wsMsg
			}

		case "MY_ATTENDANCE":
			if c.role != "student" {
				errMsg := wsError{
					Event: "ERROR",
					Data: WsErrorData{
						Message: "Forbidden, student event only",
					},
				}
				c.send <- errMsg
			} else if ActiveSession.StartedAt == "" {
				errMsg := &wsError{
					Event: "Event error",
					Data: WsErrorData{
						Message: "No active attendance sessions",
					},
				}
				c.send <- *errMsg
			} else {

				status := ""

				if value, ok := ActiveSession.AttendanceStatus[c.id]; ok {
					status = value
				} else {
					status = "not yet update"
				}

				wsMsg := WsMyAttendance{
					Event: "MY_ATTENDANCE",
					Data: WsMyAttendanceData{
						Status: status,
					},
				}
				c.send <- wsMsg
			}
		case "DONE":
			if c.role != "teacher" {
				c.send <- wsError{Event: "ERROR", Data: WsErrorData{Message: "Forbidden, teacher event only"}}
			} else if ActiveSession.StartedAt == "" {
				c.send <- wsError{Event: "ERROR", Data: WsErrorData{Message: "No active attendance session"}}
			} else {

				present := 0
				absent := 0
				total := 0

				collection := db.Database("attendance").Collection("class")
				//get all students from class id

				filter := bson.M{
					"_id": ActiveSession.ClassID,
				}
				var class data.Class
				err := collection.FindOne(context.Background(), filter).Decode(&class)
				if err != nil {
					panic(err)
				}

				//mark the students who havent joined absent
				//iterate through the studentids array, if it doesnt exist in the hashmap then add and status = absent

				for _, v := range class.StudentIDs {
					if _, ok := ActiveSession.AttendanceStatus[v.Hex()]; ok {
						continue
					} else {
						ActiveSession.AttendanceStatus[v.Hex()] = "absent"
					}
				}

				//iterate through the hashmap and for each record add an entry into attendance collection

				for k, v := range ActiveSession.AttendanceStatus {

					studentId, _ := bson.ObjectIDFromHex(k)

					Record := &data.Attendance{
						ID:        bson.NewObjectID(),
						ClassID:   ActiveSession.ClassID,
						StudentID: studentId,
						Status:    v,
					}

					_, err := db.Database("attendance").Collection("records").InsertOne(context.Background(), &Record)
					if err != nil {
						panic(err)
					}

				}

				//final summary and send

				for _, v := range ActiveSession.AttendanceStatus {
					switch v {
					case "present":
						present++
					case "absent":
						absent++
					}
				}
				total = present + absent

				//persist class data to db

				done := WsDone{
					Event: "EVENT",
					Data: WsDoneData{
						Message: "Attendance Persisted",
						Present: present,
						Absent:  absent,
						Total:   total,
					},
				}

				Message := &Message{
					ClientID: c.id,
					Type:     "DONE",
					Text:     done,
				}

				c.hub.broadcast <- Message
			}

		default:
			errMsg := wsError{
				Event: "Event error",
				Data: WsErrorData{
					Message: "invaid format",
				},
			}

			c.send <- errMsg
		}

	}

}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte("err occured man"))
				return
			}

			byteA, _ := json.Marshal(msg)

			err := c.conn.WriteMessage(websocket.TextMessage, byteA)
			if err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		}
	}
}
