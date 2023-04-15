package user_controllers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go_r5/main/consts"
	"go_r5/main/db"
	"go_r5/main/models/data_model"
	"go_r5/main/models/response"
	"go_r5/main/service"
	"go_r5/main/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
)

func Register(c *gin.Context) {
	newUser := data_model.User{}
	name := c.PostForm("user_name")
	password := c.PostForm("password")

	//检查重名
	var searchOutcome []data_model.User
	searchOutcome, err := service.SearchUser(name, 0)
	if fmt.Sprintf("%v", err) != "record not found" && err != nil {
		c.JSON(502, response.ServerFailErrorResponse([]error{err}))
		abortNLog(c, "Databases search unavailable", err)
		return
	} else if len(searchOutcome) != 0 {
		c.JSON(200, response.UsrInvldInptStringDataResponse("Used user name?", []error{err}))
		abortNLog(c, "Used user name", err)
		c.Abort()
		return
	}

	//空输入
	if name == "" || password == "" {
		c.JSON(500, response.UsrInvldInptStringDataResponse("Empty password / username!", []error{}))
		abortNLog(c, "Illegal name or password", errors.New("empty password / username"))
		c.Abort()
		return
	}

	//创建id
	userId, err := uuid.NewUUID()
	if err != nil {
		c.JSON(502, response.ServerFailErrorResponse([]error{err}))
		abortNLog(c, "Uuid generation failed ", err)
		c.Abort()
		return
	}

	//加密密码
	encryptedPasswordBytes := utils.EncryptPassword(c, password)
	if encryptedPasswordBytes == nil {
		return
	}

	newUser.Uid = userId.String()

	//生成用户数据结构
	newUser.EncryptedPassword = string(encryptedPasswordBytes)
	newUser.Name = name

	err = service.SaveUser(&newUser, 1)
	if err != nil {
		return
	}

	c.JSON(200, response.OkStringDataResponse("new user created"))
	log.Printf("new user registration success! userId: %s, name: %s", newUser.Uid, newUser.Name)
}

func Login(c *gin.Context) {
	name := c.Query("user_name")
	password := c.Query("password") //搜索用户

	searchOutcome, err := service.SearchUser(name, 0)

	if fmt.Sprintf("%v", err) == "record not found" {
		//无此用户：
		c.JSON(500, response.UsrInvldInptStringDataResponse("No such user", []error{err}))
		abortNLog(c, "No such user for name", name)
		return
	}

	if err != nil {
		c.JSON(502, response.ServerFailErrorResponse([]error{err}))
		abortNLog(c, "User search error", err)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(searchOutcome[0].EncryptedPassword), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		//错误密码：
		c.JSON(500, response.UsrInvldInptStringDataResponse("wrong password", []error{err}))
		abortNLog(c, "Wrong password: user", name)
		c.Abort()
		return
	}

	if err == nil {
		// 签发 token
		token, err := utils.Sign(consts.ACCESS_MODE().USER_MODE, searchOutcome[0].Uid, searchOutcome[0].Name)
		if err != nil {
			c.JSON(502, response.ServerFailErrorResponse([]error{err}))
			c.Abort()
			return
		}

		c.JSON(200, response.JsonDataResponse{
			Status: 200,
			Data:   gin.H{"token": token, "userId": searchOutcome[0].Uid},
			Msg:    "Login success",
			Error:  "",
		})
		abortNLog(c, "Login success: user", name)
	}
}

func SetPassword(c *gin.Context) {
	uid := c.Query("user_id")
	oldPas := c.Query("old_password")
	newPas := c.Query("new_password")

	searchOutcome, err := service.SearchUser(uid, 1)
	if searchOutcome == nil {
		return
	}

	//对比数据库旧密码和输入的旧密码：
	err = bcrypt.CompareHashAndPassword([]byte(searchOutcome[0].EncryptedPassword), []byte(oldPas))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		//错误密码：
		c.JSON(500, response.UsrInvldInptStringDataResponse("wrong password", []error{err}))
		abortNLog(c, "Wrong password: user", searchOutcome[0].Uid)
		return
	}
	if err != nil {
		c.JSON(502, response.ServerFailErrorResponse([]error{err}))
		abortNLog(c, "Error comparing password", err)
		return
	} else {
		searchOutcome[0].EncryptedPassword = string(utils.EncryptPassword(c, newPas))
		err = service.SaveUser(&searchOutcome[0], 1)
		if err != nil {
			return
		}

		c.JSON(200, response.OkStringDataResponse("password changed"))
		abortNLog(c, "change password success: user", searchOutcome[0].Uid)
	}
}

func SearchUser(c *gin.Context) { // 模糊搜索
	content := c.Query("search_content")

	var resultList []data_model.User
	var result *gorm.DB

	result = db.SqlDb.Model(&data_model.User{}).Where("name LIKE ?", "%"+content+"%").Find(&resultList)

	if result.Error != nil && fmt.Sprintf("%v", result.Error) != "record not found" {
		c.JSON(502, response.ServerFailErrorResponse([]error{result.Error}))
		abortNLog(c, "Search items failed", result.Error)
		return
	}
	makeSuccessUserListRsp(c, resultList)
	log.Printf("Search user success!")
}

// Connect 升级一个http链接to ws
func Connect(c *gin.Context) {
	uid := c.Query("user_id")
	service.UpgradeConn(c, uid)
}

func GetConversation(c *gin.Context) {
	uid1 := c.Query("user_id")
	uid2 := c.Query("target_id")
	msgList, err := service.GetConversationList(uid1, uid2)
	if err != nil {
		abortNLog(c, "GetConversation err", err)
		return
	}
	makeSuccessMsgListRsp(c, msgList)
}

func GetAllMessage(c *gin.Context) {
	uid := c.Query("user_id")
	msgList, err1, err2 := service.GetAllMessages(uid)
	if err1 != nil || err2 != nil {
		abortNLog(c, "GetConversation err", []error{err1, err2})
		return
	}
	makeSuccessMsgListRsp(c, msgList)
}

func GetGroupMessages(c *gin.Context) {
	uid := c.Query("user_id")
	msgList, err := service.GetGroupMsgs(uid)
	if err != nil {
		abortNLog(c, "GetConversation err", err)
		return
	}
	makeSuccessMsgListRsp(c, msgList)
}

func GetFriendList(c *gin.Context) {
	uid := c.Query("user_id")
	resultList, err := service.GetAllFriendList(uid)
	if err != nil {
		abortNLog(c, "Unable to fetch friend list.", err)
	}
	makeSuccessUserListRsp(c, resultList)
}

func GetGroupMemberList(c *gin.Context) {
	uid := c.Query("user_id")
	gid := c.Query("group_id")
	if service.IsThisGuyInGroup(uid, gid) {
		resultList, err := service.GetAllGroupMembers(gid)
		if err != nil {
			abortNLog(c, "Unable to fetch friend list.", err)
		}
		makeSuccessUserListRsp(c, resultList)
	} else {
		abortNLog(c, "GetGroupMemberList not a member!", errors.New("GetGroupMemberList not a member"))
	}
}

func GetAuth(c *gin.Context) {
	uid := c.Query("user_id")
	auth, err := service.GenerateAuth(uid)
	if err != nil {
		abortNLog(c, "GenerateAuth failed", err)
	}
	c.JSON(200, response.OkStringDataResponse(auth))
}

func AddFriend(c *gin.Context) {
	uid := c.Query("user_id")
	fid := c.Query("friend_id")
	auth := c.Query("auth")
	isSuccess, err := service.AddFriend(auth, uid, fid)
	if isSuccess {
		c.JSON(200, response.OkStringDataResponse("add friend success!"))
	} else {
		abortNLog(c, "unauthorized friend request / err adding friend", err)
	}
}

func GetGroupAuth(c *gin.Context) {
	uid := c.Query("user_id")
	gid := c.Query("group_id")
	auth, err := service.GenerateGroupAuth(uid, gid)
	if err != nil {
		abortNLog(c, "GenerateGroupAuth failed", err)
	}
	c.JSON(200, response.OkStringDataResponse(auth))
}

// ---------------------- utils
func abortNLog[T any](c *gin.Context, logString string, err T) {
	c.Abort()
	log.Printf(logString+"err/hint: %v", err)
}

func makeSuccessMsgListRsp(c *gin.Context, msgList []data_model.Message) {
	c.JSON(200, response.MsgListResponse{
		Data:   msgList,
		Error:  "",
		Msg:    "success!",
		Status: 200,
	})
	c.Next()
}

func makeSuccessUserListRsp(c *gin.Context, resultList []data_model.User) {
	c.JSON(200, response.UserListResponse{
		Status: 200,
		Data:   resultList,
		Msg:    "Search users success!",
		Error:  "",
	})
	c.Next()
}
