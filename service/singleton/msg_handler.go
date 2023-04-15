package singleton

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go_r5/main/db"
	"go_r5/main/models/data_model"
	"golang.org/x/net/context"
	"gorm.io/gorm"
	"log"
	"strconv"
	"time"
)

type msgHandler struct {
	isRunning        bool
	connectionMap    map[string]*websocket.Conn
	messageBufferChn chan data_model.Message
}

const (
	connection = 0
	stringData = 1
	message    = 2
)

type bundle struct {
	infoType             int
	bundleStringData     *string
	bundleMessageData    *data_model.Message
	bundleConnectionData *websocket.Conn
}

// 这些个玩意的作用域有点大，懒得改了
var recvInChn chan bundle = make(chan bundle, 10)
var recvOutChn chan bundle = make(chan bundle, 10)
var sendInChn chan bundle = make(chan bundle, 10)
var sendOutChn chan bundle = make(chan bundle, 10)

// 好原始啊，我爱 kotlin，下面是getter和setter之类的，虽然kotlin好像也没什么用：
func (h *msgHandler) AddConnection(uid string, conn *websocket.Conn) {
	h.connectionMap[uid] = conn
	recvInChn <- bundle{
		infoType:             connection,
		bundleStringData:     &uid,
		bundleConnectionData: conn,
	}
}
func (h *msgHandler) GetConnection(uid string) *websocket.Conn {
	return h.connectionMap[uid]
}
func (h *msgHandler) DelConnection(uid string) {
	delete(h.connectionMap, uid)
}
func (h *msgHandler) GetMessageBufferChn() *chan data_model.Message {
	//由于有很多携程，这个可能调取次数会很高
	return &h.messageBufferChn
}
func (h *msgHandler) GetConnList() map[string]*websocket.Conn {
	var newMap map[string]*websocket.Conn
	for k, v := range h.connectionMap {
		newMap[k] = v
	}
	return newMap
}

// Run 开始处理收到的消息队列，一开始是没有链接的
func (h *msgHandler) Run() {
	if !h.isRunning {
		h.isRunning = true
		go func(recvInChn *chan bundle, recvOutChn *chan bundle, sendInChn *chan bundle, sendOutChn *chan bundle, handler *msgHandler) {
			//todo: 尝试启动五对协程 - 有点少？：
			for i := 0; i < 5; i++ {
				go recvMsgs(recvInChn, recvOutChn)
				go sendMsgs(sendInChn, sendOutChn)
			}
			var redisSaveChn = make(chan data_model.Message, 50) //todo: 这里应该能够自定义才行
			for {
				//处理bundle
				//recvBundle := <-*recvOutChn
				//handleRecvBundle(recvBundle)
				sendBundle := <-*sendOutChn
				handleSendBundle(sendBundle, &redisSaveChn)
				//五秒同步数据库，bundle这里暂时没用，可以不用考虑
				time.Sleep(5 * time.Second)
				syncDb(&redisSaveChn)
			}
		}(&recvInChn, &recvOutChn, &sendInChn, &sendOutChn, h)
	} else {
		panic("msgHandler already running")
	}
}
func handleSendBundle(sendBundle bundle, redisSaveChn *chan data_model.Message) {
	switch {
	case sendBundle.infoType == message:
		*redisSaveChn <- *sendBundle.bundleMessageData
	}
}
func handleRecvBundle(recvBundle bundle) {

}

// 启动之后会自动运行的函数
func syncDb(redisSaveChn *chan data_model.Message) {

	ctx := context.Background()
	for {
		select {
		case msg := <-*redisSaveChn:
			err := db.RedisCli.HSet(ctx, msg.TargetId, "from_id", msg.FromId, "msg_type", msg.MsgType, "content", msg.Content).Err()
			if err != nil {
				log.Printf("syncDb redis write err: %v", err)
				panic(err)
			}
		default:
			break
		}
	}
}

func recvMsgs(recvInChn *chan bundle, recvOutChn *chan bundle) {
	go func() {
		for {
			select {
			case bundle := <-*recvInChn:
				recvDealWithBundle(bundle)
				//拿到外面发的链接，进行一个循环接收先
			}
		}
	}()
}
func recvDealWithBundle(bundle bundle) {
	if bundle.infoType == connection {
		//新的链接，建立一个协程进行监听
		go func(conn *websocket.Conn, uid *string) {
			if conn == nil || uid == nil {
				//我不觉得会发生
				log.Printf("nil ws.conn or *string when recvDealWithBundle(connection)?")
			} else {
				listenConn(conn, *uid)
			}
		}(bundle.bundleConnectionData, bundle.bundleStringData)
	}
}
func listenConn(conn *websocket.Conn, uid string) {
	//拿到一个刚刚升级的ws链接
	//不再接收就关闭连接：
	defer func(conn *websocket.Conn, uid string) {
		msgHandler := GetHolder().GetMsgHandler()
		err := conn.Close()
		if err != nil {
			log.Printf("err when defering websocket connection! err: %v", err)
		}
		msgHandler.DelConnection(uid)
	}(conn, uid)

	//循环接收
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			//客户端关闭连接
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("custom connction closed: %v", err)
				break
			} else {
				//其他错误，目前是一有错误就断开链接
				log.Printf("recvProc ReadMessage err: %v", err)
				break
			}
		}

		//处理消息，妈的这是第几层协程？三层？
		go func(data []byte, msgBufferChn *chan data_model.Message) {
			//这里获取到的是message对象结构的json
			var msg map[string]string
			err = json.Unmarshal(data, &msg)
			if err != nil {
				log.Printf("recvProc Unmarshal json err: %v", err)
				return
			}
			if err != nil {
				//todo： uuid拉不出来就毁灭吧
				panic("unable to generate uuid for msg, really?")
			}
			message := data_model.Message{
				SendTime: strconv.FormatInt(time.Now().Unix(), 10),
				FromId:   msg["from_id"],
				TargetId: msg["target_id"],
				MsgType:  msg["msg_type"],
				Content:  msg["content"],
			}
			// 权限判断
			if message.MsgType == "1" {
				err := db.SqlDb.Model(&data_model.Contact{}).Where("owner_id = ? AND target_id = ?", message.FromId, message.TargetId).First(nil).Error
				if err == gorm.ErrRecordNotFound {
					log.Println("sending, access not granted. From: ", message.FromId, "To: ", message.TargetId)
					return
				}
			} else if message.MsgType == "2" {
				err := db.SqlDb.Model(&data_model.Contact{}).Where("owner_id = ? AND target_id = ?", message.TargetId, message.FromId).First(nil).Error //群聊的ID是目标，但是表里面的字段是owner
				if err == gorm.ErrRecordNotFound {
					log.Println("group sending, access not granted. From: ", message.FromId, "To: ", message.TargetId)
					return
				}
			}

			id, err := uuid.NewRandom()
			message.Id = id.String()
			msg["id"] = id.String()
			//加入 redis 发送历史消息记录
			err = db.AddMsgToSendList(message.FromId, msg)
			if err != nil {
				log.Println("Unable to put message into redis! err: ", err)
			}

			//发到messagebuffer等處理
			*msgBufferChn <- message

		}(data, GetHolder().GetMsgHandler().GetMessageBufferChn())
	}
}

func sendMsgs(sendInChn *chan bundle, sendOutChn *chan bundle) {
	go func() {
		for {
			select {
			case bundle := <-*sendInChn:
				sendDealWithBundle(bundle)
			//拿到外面发的链接，进行一个循环接收先

			case msg := <-*GetHolder().GetMsgHandler().GetMessageBufferChn():
				go sendMsg(msg)
			}
		}
	}()
}
func sendDealWithBundle(bundle bundle) {
	//目前没什么要做的
}
func sendMsg(msg data_model.Message) {
	//检测类型，GC/DM？
	var targetList = make([]string, 0)
	if msg.MsgType == "1" {
		//DM
		targetList = append(targetList, msg.TargetId)
	} else if msg.MsgType == "2" {
		//GC
		var tempList = make([]data_model.Contact, 0)
		db.SqlDb.Model(&data_model.Contact{}).Where("owner_id = 1 AND (type = 2 OR type = 3)", msg.TargetId, "2", "3").Find(&tempList)
		for _, contact := range tempList {
			db.RedisCli.RPush(context.Background(), "recv:"+contact.TargetId)
			targetList = append(targetList, contact.TargetId)
		}
	}

	for _, targetId := range targetList {
		go func(targetId string) {
			targetConn := GetHolder().GetMsgHandler().GetConnection(targetId)
			if targetConn != nil {
				//在线才能发送
				data, err := json.Marshal(msg)
				if err != nil {
					log.Printf("sendMsg json.Marshal(msg) err: %v ", err)
					return
				}
				err = targetConn.WriteMessage(websocket.TextMessage, data)
				if err != nil {
					log.Printf("sendMsg targetConn.WriteMessage err: %v ", err)
					return
				}
			}
			err := db.AddMsgToRecvList(targetId, msg.Id)
			if err != nil {
				log.Printf("sendMsg targetConn.WriteMessage err: %v ", err)
				return
			}
		}(targetId)
	}
}
