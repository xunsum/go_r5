package response

import "github.com/gin-gonic/gin"

type JsonDataResponse struct {
	Status int    `json:"status"`
	Data   gin.H  `json:"data"`
	Msg    string `json:"msg"`
	Error  string `json:"error"`
}
