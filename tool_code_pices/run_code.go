package tool_code_pices

import (
	"encoding/json"
	"fmt"
	"go_r5/main/models/data_model"
	"strconv"
	"time"
)

func RunCode() {
	fmt.Println("RunCode ---------------------------")
	fmt.Println("测试Marshal/Un...")

	msg := data_model.Message{
		SendTime: strconv.FormatInt(time.Now().Unix(), 10),
		FromId:   "abcd1234",
		TargetId: "4321dcba",
		MsgType:  "1",
		Content:  "Test content",
	}

	fmt.Println("msg_ori: ", msg)

	data, err := json.Marshal(msg)
	fmt.Printf("json.Marshal(msg) err: %v\n", err)
	fmt.Printf("json.Marshal(msg) data: %v\n", data)

	var msgBack data_model.Message
	err = json.Unmarshal(data, &msgBack)
	fmt.Printf("json.Unmarshal(data, &msg_back) err: %v\n", err)
	fmt.Printf("json.Unmarshal(data, &msg_back) msg_back: %v\n", msgBack)

	panic("RunCode over!----------------------")
}
