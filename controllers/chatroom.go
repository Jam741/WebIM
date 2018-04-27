package controllers

import (
	"WebIM/models"
	"time"
	"github.com/gorilla/websocket"
	"container/list"
	"github.com/astaxie/beego"
)

/*数据管理相关的函数*/

type Subscription struct {
	Archive []models.Event      // All the events from the archive. 来自存档的所有事件。
	New     <-chan models.Event // New events coming in. 新事件进来
}

func newEvent(ep models.EventType, user, msg string) models.Event {
	return models.Event{Type: ep, User: user, Timestamp: int(time.Now().Unix()), Content: msg}
}

func Join(user string, ws *websocket.Conn) {
	subscribe <- Subscriber{user, ws}
}

func Leave(user string) {
	unsubscribe <- user
}

type Subscriber struct {
	Name string
	Conn *websocket.Conn // Only for WebSocket users; otherwise nil. 只对WebSocket用户;否则nil。
}

var (
	// Channel for new join users. 新的连接用户通道。
	subscribe = make(chan Subscriber, 10)
	// Channel for exit users. 用户退出渠道。
	unsubscribe = make(chan string, 10)
	// Send events here to publish them. 发送事件来发布它们。
	publish = make(chan models.Event, 10)
	// Long polling waiting list.
	waitingList = list.New()
	subscribers = list.New()
)

// This function handles all incoming chan messages. 此函数处理所有传入的chan消息。
func chatroom() {
	for {
		select {
		case sub := <-subscribe:
			if !isUserExist(subscribers, sub.Name) { //用户没有退出
				// Add user to the end of list.
				subscribers.PushBack(sub)
				// Publish a JOIN event. 发布一个连接事件
				beego.Info("New user:", sub.Name, ";WebSocket:", sub.Conn != nil)
			} else {
				beego.Info("Old user:", sub.Name, ";WebSocket:", sub.Conn != nil)
			}
		case event := <-publish:
			for ch := waitingList.Back(); ch != nil; ch = ch.Prev() {
				ch.Value.(chan bool) <- true
				waitingList.Remove(ch)
			}
			broadcastWebSocket(event)
			models.NewArchive(event)
			if event.Type == models.EVENT_MESSAGE {
				beego.Info("Message from", event.User, ";Content", event.Content)
			}
		case unsub := <-unsubscribe:
			for sub := subscribers.Front(); sub != nil; sub.Next() {
				if sub.Value.(Subscriber).Name == unsub {
					subscribers.Remove(sub)
					// Clone connection
					ws := sub.Value.(Subscriber).Conn
					if ws != nil {
						ws.Close()
						beego.Error("WebSocket closed:", unsub)
					}
					publish <- newEvent(models.EVENT_LEAVE, unsub, "") // Publish a LEAVE event.
					break
				}
			}
		}
	}
}

func init() {
	go chatroom()
}

func isUserExist(subscribers *list.List, user string) bool {
	for sub := subscribers.Front(); sub != nil; sub.Next() {
		if sub.Value.(Subscriber).Name == user {
			return true
		}
	}
	return false
}
