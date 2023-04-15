package utils

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go_r5/main/models/response"
	"golang.org/x/crypto/bcrypt"
	"log"
)

func EncryptPassword(c *gin.Context, password string) []byte {
	encryptedPasswordBytes, err2 := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err2 != nil {
		c.JSON(502, response.StringDataResponse{
			Status: 502,
			Data:   "",
			Msg:    "Having problem generating user info.",
			Error:  fmt.Sprintf("EncryptedPswd generation failed, error: %v", err2),
		})
		log.Printf("EncryptedPswd generation failed, error: %v", err2)
		c.Abort()
		return nil
	}
	return encryptedPasswordBytes
}
