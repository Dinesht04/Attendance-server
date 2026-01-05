package server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dinesht04/ws-attendance/util"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var upgrader = websocket.Upgrader{}

type Data struct {
	StudentID *string `json:"studentID"`
	Status    *string `json:"status"`
}

type WsReq struct {
	Event string  `json:"event"`
	Data  *Data   `json:"data"`
	Error *string `json:"err"`
}

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
			send: make(chan WsReq),
		}

		client.hub.register <- client

		go client.writePump()
		go client.readPump()

		// for {
		// 	mt, message, err := c.ReadMessage()
		// 	if err != nil {
		// 		util.InternalServerError(ctx, err, "message reading err")
		// 		break
		// 	}

		// 	valid := json.Valid(message)
		// 	if !valid {
		// 		fmt.Println(valid)
		// 		err = c.WriteMessage(mt, []byte("Invalid json"))
		// 		if err != nil {
		// 			util.InternalServerError(ctx, err, "message sending err")
		// 			break
		// 		}
		// 		continue
		// 	}
		// 	data := WsReq{}

		// 	err = json.Unmarshal(message, &data)
		// 	if err != nil {
		// 		util.InternalServerError(ctx, err, "json unmarshalling err")

		// 		err = c.WriteMessage(mt, []byte("Invalid json"))
		// 		if err != nil {
		// 			util.InternalServerError(ctx, err, "message sending err")
		// 			break
		// 		}

		// 		continue
		// 	}

		// 	switch data.Event {
		// 	case "ATTENDANCE_MARKED":
		// 		fmt.Println("attendance marked")
		// 	case "TODAY_SUMMARY":
		// 		fmt.Println("today summary")
		// 	case "test":
		// 		SocketList.SendMessage()
		// 	case "MY_ATTENDANCE":
		// 		fmt.Println("my attendance")
		// 	case "DONE":
		// 		fmt.Println("dnoe")
		// 	default:
		// 		err = c.WriteMessage(mt, []byte("Invalid event"))
		// 		if err != nil {
		// 			util.InternalServerError(ctx, err, "message sending err")
		// 		}
		// 	}

		// 	err = c.WriteMessage(mt, message)
		// 	if err != nil {
		// 		util.InternalServerError(ctx, err, "message sending err")
		// 		break
		// 	}
		// }
	}
}

func (c *Client) readPump() {
	defer func() {
		c.conn.Close()
		c.hub.unregister <- c
	}()

	for {
		mt, msg, err := c.conn.ReadMessage()
		if err != nil {
			fmt.Println(err, "reading msg error")
		}

		valid := json.Valid(msg)

		if !valid {
			fmt.Println(valid)
			err = c.conn.WriteMessage(mt, []byte("Invalid json"))
			if err != nil {
				fmt.Println(err, "message sending err invalid json")
				break
			}
			continue
		}
		data := WsReq{}

		err = json.Unmarshal(msg, &data)
		if err != nil {
			fmt.Println(err, "json unmarshalling err")

			err = c.conn.WriteMessage(mt, []byte("Invalid json"))
			if err != nil {
				fmt.Println(err, "message sending err json umarshalling")
				break
			}

			continue
		}
		message := &Message{
			ClientID: c.id,
			Text:     data,
		}

		//this is where we handle everything

		switch data.Event {
		case "ATTENDANCE_MARKED":

			c.hub.broadcast <- message
		case "TODAY_SUMMARY":
			fmt.Println("today summary")
		case "test":
			fmt.Println("test")
		case "MY_ATTENDANCE":
			fmt.Println("my attendance")
		case "DONE":
			c.hub.broadcast <- message
		default:

			err := "unknown event"

			errMsg := &WsReq{
				Event: "Event error",
				Error: &err,
			}

			c.send <- *errMsg
		}

		//if event type is broadcast

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
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
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
