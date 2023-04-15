package response

import (
	"go_r5/main/models/data_model"
)

type UserListResponse struct {
	Data   []data_model.User `json:"data"`
	Error  string            `json:"error"`
	Msg    string            `json:"msg"`
	Status int64             `json:"status"`
}
