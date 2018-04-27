package controllers

import (
	"github.com/gorilla/websocket"
	"net/http"
	"github.com/astaxie/beego"
	"WebIM/models"
	"encoding/json"
)

/* WebSocket 模式的控制器和方法*/

// WebSocketController handles WebSocket requests.
type WebSocketController struct {
	baseController
}

// Get method handles GET requests for WebSocketController.
func (this *WebSocketController) Get() {
	// Safe check.
	uname := this.GetString("uname")
	if len(uname) == 0 {
		this.Redirect("/", 302)
		return
	}

	this.TplName = "websocket.html"
	this.Data["IsWebSocket"] = true
	this.Data["UserName"] = uname
}

// Join method handles WebSocket requests for WebSocketController.
func (this *WebSocketController) Join() {
	uname := this.GetString("uname")
	if len(uname) == 0 {
		this.Redirect("/", 302)
		return
	}

	ws, err := websocket.Upgrade(this.Ctx.ResponseWriter, this.Ctx.Request, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(this.Ctx.ResponseWriter, "Not websocket handshake", 400)
		return
	} else if err != nil {
		beego.Error("Cannot setup WebSocket connection", err)
		return
	}

	// Join chat room.
	Join(uname, ws)
	defer Leave(uname)

	// Message receive loop.
	for {
		_, p, err := ws.ReadMessage()
		if err != nil {
			return
		}
		publish <- newEvent(models.EVENT_MESSAGE, uname, string(p))
	}
}

// broadcastWebSocket broadcasts messages to WebSocket users.
func broadcastWebSocket(event models.Event) {
	data, err := json.Marshal(event)
	if err != nil {
		beego.Error("Fail to marshal event:", err)
		return
	}

	for sub := subscribers.Front(); sub != nil; sub.Next() {
		// Immediately send event to WebSocket users. 立即向WebSocket用户发送事件。
		ws := sub.Value.(Subscriber).Conn
		if ws != nil {
			if ws.WriteMessage(websocket.TextMessage, data) != nil {
				//User disconnected
				unsubscribe <- sub.Value.(Subscriber).Name
			}
		}

	}
}
