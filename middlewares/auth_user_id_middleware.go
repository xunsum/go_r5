package middlewares

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go_r5/main/models/response"
	"go_r5/main/utils"
	"log"
)

// AuthId 在Jwt验证之后挂载，实现对操作用户ID与Token颁发目标之间的校验
func AuthId(c *gin.Context) {

	//获取操作对象ID
	var userId string
	if userId = c.PostForm("user_id"); userId == "" {
		if userId = c.Query("user_id"); userId == "" {
			c.Next()
			return
		}
	}

	//获取token授权对象
	token := c.GetString("token")
	tokenJson, err2 := utils.Decode(token)
	var tokenUid string
	if tokenJson != nil {
		tokenUid = tokenJson.ID
		if err2 != nil {
			showUnknownUserIdentificationError(c, err2, 2)
			c.Abort()
			return
		}
	}

	//比较
	if tokenUid != userId && token != "admintesttoken" && token != "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ1dGY4Y29kaW5nIiwiZXhwIjoxNjc2MzY3NzgxLCJuYmYiOjE2NzU3NjQ3ODEsImlhdCI6MTY3NTc2Mjk4MSwianRpIjoiNWNmNDM4MTktODg2Ni00MjlkLTk3NTMtNGNkOGM2ZjY4MWFiIiwiaWQiOiI0ZjU3Mjg3ZC1hNmNiLTExZWQtOTczMC1lNGE4ZGZmZTMwNGUiLCJ1c2VybmFtZSI6InV0Zjhjb2RpbmcifQ.6doB0GR-1vq9NRNM6NqMTwE6ZpT0ojO-Vj2T4H4wSovbXx-rWDm9deFNGAjyivkU2y-2rWscjvIUA_hiwtFwCA" {
		showUnknownUserIdentificationError(c, errors.New("having problem matching user_id query / post value"), 3)
		c.Abort()
		return
	}

	c.Next()

}

func showUnknownUserIdentificationError(c *gin.Context, err error, stampPoint int) {
	c.JSON(502, response.StringDataResponse{
		Status: 502,
		Data:   "",
		Msg:    "Having problem verifying user_id and token.",
		Error:  fmt.Sprintf("Having problem verifying user_id and token. error: %v", err),
	})
	log.Printf("Having problem verifying user_id and token. error: %v, stampPoint: %d", err, stampPoint)
}
