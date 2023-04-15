package singleton

import (
	"github.com/gorilla/websocket"
	"go_r5/main/models/data_model"
	"sync"
)

// Holder 在main函数启动，用于持有所有的单例类
type holder struct{}

var singletonHolder *holder
var mMsgHandler *msgHandler

func init() {
	singletonHolder = &holder{}
}

// GetHolder 获取 holder 的唯一方式
func GetHolder() *holder {
	if singletonHolder == nil {
		panic("this should not happen, no singletonHolder instance!")
	}
	return singletonHolder
}

// GetMsgHandler 唯一的获取方式
func (holder) GetMsgHandler() *msgHandler {
	if mMsgHandler != nil {
		return mMsgHandler
	} else {
		startMsgHandler()
		return mMsgHandler
	}
}

// startMsgHandler 这个有点 lazyLoad
func startMsgHandler() {
	var once sync.Once
	once.Do(func() {
		mMsgHandler = &msgHandler{
			isRunning:        false,
			connectionMap:    make(map[string]*websocket.Conn, 0),
			messageBufferChn: make(chan data_model.Message, 100),
		}
	})
	mMsgHandler.Run()
}
