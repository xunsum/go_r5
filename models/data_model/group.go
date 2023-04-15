package data_model

type Group struct {
	Name    string
	OwnerId string
	Type    int
}

func (Table *Group) TableName() string {
	return "groups"
}
