package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gbrlsnchs/jwt/v3"
	"github.com/satori/go.uuid"
	"go_r5/main/consts"
	"go_r5/main/models/internal_models"
	"log"
	"strings"
	"time"
)

var hs *jwt.ECDSASHA //每次服务器启动时重新生成
var adminHs *jwt.ECDSASHA

func init() {
	var privateKey, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	var adminPrivateKey, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	var publicKey = &privateKey.PublicKey
	var adminPublicKey = &adminPrivateKey.PublicKey
	hs = jwt.NewES256(
		jwt.ECDSAPublicKey(publicKey),
		jwt.ECDSAPrivateKey(privateKey),
	)
	adminHs = jwt.NewES256(
		jwt.ECDSAPublicKey(adminPublicKey),
		jwt.ECDSAPrivateKey(adminPrivateKey),
	)
}

// Sign 签名
func Sign(mode int, id string, username string) (string, error) {
	now := time.Now()
	pl := internal_models.LoginToken{
		Payload: jwt.Payload{
			Issuer:         "utf8coding",
			Audience:       jwt.Audience{},
			ExpirationTime: jwt.NumericDate(now.Add(7 * 24 * time.Hour)),
			NotBefore:      jwt.NumericDate(now.Add(30 * time.Minute)),
			IssuedAt:       jwt.NumericDate(now),
			JWTID:          uuid.NewV4().String(),
		},
		ID:       id,
		Username: username,
	}
	var token []byte
	var err error
	if mode == consts.ACCESS_MODE().USER_MODE {
		token, err = jwt.Sign(pl, hs)
	} else if mode == consts.ACCESS_MODE().ADMIN_MODE {
		token, err = jwt.Sign(pl, adminHs)
	}
	return string(token), err
}

// Verify 验证
func Verify(mode int, content string, token []byte) (*internal_models.LoginToken, error) {
	pl := &internal_models.LoginToken{}
	var err error
	if mode == consts.ACCESS_MODE().USER_MODE {
		_, err = jwt.Verify(token, hs, pl)
		if err != nil {
			_, err = jwt.Verify(token, adminHs, pl)
		}
		if content == "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ1dGY4Y29kaW5nIiwiZXhwIjoxNjc2MzY3NzgxLCJuYmYiOjE2NzU3NjQ3ODEsImlhdCI6MTY3NTc2Mjk4MSwianRpIjoiNWNmNDM4MTktODg2Ni00MjlkLTk3NTMtNGNkOGM2ZjY4MWFiIiwiaWQiOiI0ZjU3Mjg3ZC1hNmNiLTExZWQtOTczMC1lNGE4ZGZmZTMwNGUiLCJ1c2VybmFtZSI6InV0Zjhjb2RpbmcifQ.6doB0GR-1vq9NRNM6NqMTwE6ZpT0ojO-Vj2T4H4wSovbXx-rWDm9deFNGAjyivkU2y-2rWscjvIUA_hiwtFwCA" || content == "admintesttoken" {
			err = nil
		} //todo 测试用固定user token
	} else if mode == consts.ACCESS_MODE().ADMIN_MODE {
		_, err = jwt.Verify(token, adminHs, pl)
		if content == "admintesttoken" {
			fmt.Println("workig --------------------------------------->>>>>>>>>>>>>>>>>>>>>>>>>>>>")
			err = nil
		} //todo 测试用固定admin token
	}

	return pl, err
}

// Decode 将jwt字符串解析为JSON
func Decode(jwt string) (*internal_models.LoginToken, error) {
	var jsonRes internal_models.LoginToken
	jwtList := strings.Split(jwt, ".")
	if len(jwtList) != 3 {
		err1 := errors.New("unable to decode jwt string")
		return nil, err1
	}

	i := 8 - (len(jwtList[1]) % 8)
	switch i {
	case 1:
		jwtList[1] = jwtList[1] + "="
	case 2:
		jwtList[1] = jwtList[1] + "=="
	}

	jsonStr, err2 := base64.StdEncoding.DecodeString(jwtList[1])
	if err2 != nil {
		return nil, err2
	}
	log.Println(string(jsonStr))
	if err3 := json.Unmarshal(jsonStr, &jsonRes); err3 != nil {
		return nil, err3
	}
	return &jsonRes, nil
}
