package response

import "fmt"

// StringDataResponse 用于返回简单文字信息
type StringDataResponse struct {
	Status int    `json:"status"`
	Data   string `json:"data"`
	Msg    string `json:"msg"`
	Error  string `json:"error"`
}

func OkStringDataResponse(data string) StringDataResponse {
	rsp := StringDataResponse{
		Status: 200,
		Data:   data,
		Msg:    "Success!",
		Error:  "",
	}

	return rsp
}

func ServerFailErrorResponse(error []error) StringDataResponse {
	rsp := StringDataResponse{
		Status: 502,
		Data:   "",
		Msg:    "Server Failure!",
		Error:  fmt.Sprintf("error list: %v", error),
	}

	return rsp
}

func ServerFailStringDataResponse(data string, error []error) StringDataResponse {
	rsp := StringDataResponse{
		Status: 502,
		Data:   data,
		Msg:    "Server Failure!",
		Error:  fmt.Sprintf("error list: %v", error),
	}

	return rsp
}

func UsrInvldInptStringDataResponse(data string, err []error) StringDataResponse {
	rsp := StringDataResponse{
		Status: 500,
		Data:   data,
		Msg:    "Invalid request!",
		Error:  fmt.Sprintf("error list: %v", err),
	}

	return rsp
}
