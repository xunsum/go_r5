package db

import (
	"go_r5/main/models/data_model"
	"gorm.io/gorm"
	"log"
)

func GetFriendList(uid string) ([]data_model.User, error) {
	var resultList []data_model.User
	err := SqlDb.Model(&data_model.User{}).Where("owner_id = ?", uid).Find(&resultList).Error
	return resultList, err
}

func GetGroupMembers(gid string) ([]data_model.User, error) {
	var resultList []data_model.User
	err := SqlDb.Model(&data_model.User{}).Where("owner_id = ?", gid).Find(&resultList).Error
	return resultList, err
}

func IsContactExist(tid string, oid string, contactType int) bool {
	err := SqlDb.Model(&data_model.Contact{}).Where("owner_id = ? AND target_id = ? AND type = ?", oid, tid, contactType).Error
	if err == gorm.ErrRecordNotFound {
		return false
	} else if err == nil {
		return true
	} else {
		log.Println("IsContactExist err: ", err)
		return false
	}
}

func CreatContact(oid string, tid string, contactType int) error {
	err := SqlDb.Model(&data_model.Contact{}).Where("owner_id = ? AND target_id = ? AND type = ?", oid, tid, contactType).Save(data_model.Contact{
		TargetId: tid,
		OwnerId:  oid,
		Type:     contactType,
	}).Error
	return err
}
