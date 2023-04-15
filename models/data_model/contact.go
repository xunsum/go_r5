package data_model

type Contact struct {
	OwnerId  string `json:"owner_id"`
	TargetId string `json:"target_id"`
	Type     int    `json:"type"` //1-DM 2-群组群众 3-群组管理员 2与3的是群号对个人ID。1的是个人ID对个人ID，添加时注意添加双向。
}

func (Table *Contact) TableName() string {
	return "contacts"
}
