package routers

import (
	"github.com/gin-gonic/gin"
	"go_r5/main/controllers/user_controllers"
	"go_r5/main/middlewares"
)

func InitUserRouters(r *gin.Engine) {
	unAccessedUserRouters := r.Group("/welcome")
	{
		//注册
		unAccessedUserRouters.POST("/register", user_controllers.Register)
		//登录
		unAccessedUserRouters.GET("/login", user_controllers.Login)

	}

	wsRouters := r.Group("/ws")
	wsRouters.Use(middlewares.AuthJWT(0), middlewares.AuthId)
	{
		wsRouters.GET("/upgradeConn", user_controllers.Connect)
		wsRouters.GET("/conversation", user_controllers.GetConversation)
		wsRouters.GET("/allMsg", user_controllers.GetAllMessage)
		wsRouters.GET("/groupChat", user_controllers.GetGroupMessages)

	}

	userRouters := r.Group("/user")
	userRouters.Use(middlewares.AuthJWT(0), middlewares.AuthId)
	{
		userRouters.GET("/getGroupMemberList", user_controllers.GetGroupMemberList)
		userRouters.GET("/getFriendList", user_controllers.GetFriendList)
		userRouters.GET("/search", user_controllers.SearchUser)
		userRouters.PUT("/putPassword", user_controllers.SetPassword)
		userRouters.GET("/getFriendAuth", user_controllers.GetAuth)
		userRouters.GET("/getGroupAuth", user_controllers.GetGroupAuth)
		userRouters.GET("/addFriend", user_controllers.AddFriend)
	}
}
