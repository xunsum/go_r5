package data_model

// Message todo: msg实际上应该每一条带一个token的，然后在service中解析前进行验证，但是时间不够了
type Message struct {
	Id       string `json:"id"`
	SendTime string `json:"send_time"`
	FromId   string `json:"from_id"`
	TargetId string `json:"target_id"`
	MsgType  string `json:"msg_type"` //1-DM 2-G
	Content  string `json:"content"`
}

func (table *Message) TableName() string {
	return "message"
}
