package tool_code_pices

import (
	"fmt"
	"go_r5/main/db"
	"go_r5/main/models/data_model"
)

func InitDatabaseSheets() {
	err := db.SqlDb.AutoMigrate(&data_model.Contact{})
	err = db.SqlDb.AutoMigrate(&data_model.Group{})
	if err != nil {
		fmt.Println("%v", err)
	}
}
