package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go_r5/main/db"
	"go_r5/main/models/data_model"
	"go_r5/main/service/singleton"
	"gorm.io/gorm"
	"log"
	"net/http"
)

// SearchUser
// mode: Search mode, name - 0 id - 1
func SearchUser(searchContent string, mode int) ([]data_model.User, error) {
	var searchOutcome []data_model.User
	var result *gorm.DB
	if mode == 0 {
		result = db.SqlDb.Model(&data_model.User{}).Where("name = ?", searchContent).First(&searchOutcome)
	} else if mode == 1 {
		result = db.SqlDb.Model(&data_model.User{}).Where("uid = ?", searchContent).First(&searchOutcome)
	}
	if result.Error != nil {
		return nil, result.Error
	} else {
		return searchOutcome, nil
	}
}

// SaveUser
// mode: Search mode, name - 0 id - 1
func SaveUser(user *data_model.User, mode int) error {
	//存储用户信息
	var result *gorm.DB
	if mode == 1 {
		result = db.SqlDb.Model(&data_model.User{}).Where("uid = ?", user.Uid).Save(&user)
	} else if mode == 0 {
		result = db.SqlDb.Model(&data_model.User{}).Where("name = ?", user.Name).Save(&user)
	}

	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpgradeConn
// 升级链接，然后把链接丢给handler处理
func UpgradeConn(c *gin.Context, uid string) {
	var upGrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true //todo:防止跨站伪造，升级之前验证一次token
		},
	}

	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("websocket upgrade err: %v", err)
	}
	msgHandler := singleton.GetHolder().GetMsgHandler()
	if err != nil {
		log.Printf("UpgradeConn GetMsgHandler err: %v", err)
	}
	//单例类中添加链接，单例类中会自动处理消息接收和发送
	msgHandler.AddConnection(uid, conn)
}

func GetConversationList(uid1 string, uid2 string) ([]data_model.Message, error) {
	list, err := db.GetMsgsOfTwo(uid1, uid2)
	return list, err
}

func GetAllMessages(uid string) ([]data_model.Message, error, error) {
	recvList, err1 := db.GetAllRecvMsgs(uid)
	sendList, err2 := db.GetAllSendMsgs(uid)
	return append(recvList, sendList...), err1, err2
}

func GetGroupMsgs(groupId string) ([]data_model.Message, error) {
	return db.GetMsgOfGroup(groupId)
}

func GetAllFriendList(uid string) ([]data_model.User, error) {
	return db.GetFriendList(uid)
}

func GetAllGroupMembers(gid string) ([]data_model.User, error) {
	return db.GetGroupMembers(gid)
}

func IsThisGuyInGroup(uid string, gid string) bool {
	//todo 这个错误处理有点不太合理，但是问题不大
	return db.IsContactExist(uid, gid, 1)
}

// AddFriend return: (isSuccess, listOfErrors)
func AddFriend(auth string, uid1 string, uid2 string) (bool, []error) {
	isAuth1, err1 := compareAuth(auth, uid1)
	isAuth2, err2 := compareAuth(auth, uid2)
	if (isAuth1 || isAuth2) && err1 == nil && err2 == nil {
		err3 := db.CreatContact(uid1, uid2, 1)
		err4 := db.CreatContact(uid2, uid1, 1)
		return true, []error{err1, err2, err3, err4}
	} else {
		return false, []error{err1, err2}
	}
}

func GenerateAuth(uid string) (string, error) {
	auth := uuid.New().String()
	err := db.SetAuth(uid, auth)
	return auth, err
}

func GenerateGroupAuth(uid string, gid string) (string, error) {
	auth := uuid.New().String()
	if db.IsContactExist(uid, gid, 3) { //检测uid是管理员权限
		err := db.SetAuth(gid, auth)
		return auth, err
	} else {
		return "", errors.New("group auth generate access not granted")
	}
}

func compareAuth(uid string, auth string) (bool, error) {
	realAuth, err := db.GetAuth(uid)
	if err != nil {
		return false, err
	}
	return realAuth == auth, nil
}
