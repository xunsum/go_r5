package middlewares

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go_r5/main/models/response"
	"go_r5/main/utils"
	"log"
	"strings"
)

func AuthJWT(mode int) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		headerList := strings.Split(header, " ")
		if len(headerList) != 2 {
			err := errors.New("无法解析 Authorization 字段: ")
			showUnknownTokenError(c, err, 1, header)
			c.Abort()
			return
		}
		t := headerList[0]
		content := headerList[1]
		if t != "Bearer" {
			err := errors.New("认证类型错误, 当前只支持 Bearer")
			showUnknownTokenError2(c, err, 2, header, content, t)
			c.Abort()
			return
		}

		var err2 error
		_, err := utils.Verify(mode, content, []byte(content))

		if err2 != nil || err != nil {
			//测试用假token
			showUnknownTokenError(c, err, 3, header) //这里偷懒了，懒得写新的返回了，直接用通用错误返回了
			c.Abort()
			log.Printf("-------------------------------------content: %v", content)
		}

		c.Set("token", content)
		c.Next()
	}
}

func showUnknownTokenError(c *gin.Context, err error, stampPoint int, header string) {
	c.JSON(502, response.StringDataResponse{
		Status: 502,
		Data:   "",
		Msg:    "Having problem verifying token.",
		Error:  fmt.Sprintf("Having problem verifying token. error: %v", err),
	})
	log.Printf("Having problem verifying token. error: %v, stampPoint: %d, header: \n%s", err, stampPoint, header)
}

func showUnknownTokenError2(c *gin.Context, err error, stampPoint int, header string, content string, t string) {
	c.JSON(502, response.StringDataResponse{
		Status: 502,
		Data:   "",
		Msg:    "Having problem verifying token.",
		Error:  fmt.Sprintf("Having problem verifying token. error: %v", err),
	})
	log.Printf("Having problem verifying token. error: %v, stampPoint: %d, header: \n%s\ncontent: %s, t: %s", err, stampPoint, header, content, t)
}
