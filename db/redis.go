package db

import (
	"fmt"
	"github.com/juliangruber/go-intersect"
	"go_r5/main/models/data_model"
	"golang.org/x/net/context"
	"time"
)

func AddMsgToSendList(fromId string, msgDataMap map[string]string) error {
	//发件人发件目录
	if err := RedisCli.RPush(context.Background(), "send:"+fromId, msgDataMap["id"]).Err(); err != nil {
		return err
	}
	//message 数据哈希
	if err := RedisCli.HMSet(context.Background(), "msg:"+msgDataMap["id"], msgDataMap).Err(); err != nil {
		return err
	} else {
		RedisCli.Expire(context.Background(), "msg:"+msgDataMap["id"], 1*time.Hour) // 一小时过期时间 todo：过期时间参数化
	}
	return nil
}

func AddMsgToRecvList(targetId string, msgId string) error {
	//收件人收件目录
	if err := RedisCli.RPush(context.Background(), "recv:"+targetId, msgId).Err(); err != nil {
		return err
	}
	return nil
}

// GetAllSentMsgs GetAllRecvMsgs 都是通过一个ID拿到所有目前缓存的消息，应该是按收到时间排序的
func GetAllSentMsgs(uid string) ([]data_model.Message, error) {
	//先获取目录
	msgIdList, err := RedisCli.LRange(context.Background(), "send:"+uid, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	//创建一个List用于存放：
	var msgList = make([]data_model.Message, len(msgIdList))

	//循环获取message：
	itrCounter := 0
	for _, msgId := range msgIdList {
		resultMsg, err := RedisCli.HGetAll(context.Background(), "msg:"+msgId).Result()
		if err != nil {
			return nil, err
		} else if len(resultMsg) == 0 {
			//过期了，删掉这个目录里面的
			RedisCli.LRem(context.Background(), "send:"+uid, 0, msgId) //不处理错误，白兰
		} else {
			msgList[itrCounter] = data_model.Message{
				Id:       resultMsg["id"],
				SendTime: resultMsg["send_time"],
				FromId:   resultMsg["from_id"],
				TargetId: resultMsg["target_id"],
				MsgType:  resultMsg["msg_type"],
				Content:  resultMsg["content"],
			}
			itrCounter++
		}
	}
	msgList = msgList[:itrCounter]

	return msgList, nil
}

func GetAllRecvMsgs(uid string) ([]data_model.Message, error) {
	return getRecOrSendMsg(uid, true)
}
func GetAllSendMsgs(uid string) ([]data_model.Message, error) {
	return getRecOrSendMsg(uid, false)
}
func getRecOrSendMsg(uid string, isRecv bool) ([]data_model.Message, error) {
	var target string
	if isRecv {
		target = "recv:"
	} else {
		target = "send:"
	}
	//先获取目录
	msgIdList, err := RedisCli.LRange(context.Background(), target+uid, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	//创建一个List用于存放：
	var msgList = make([]data_model.Message, len(msgIdList))

	//循环获取message：
	itrCounter := 0
	for _, msgId := range msgIdList {
		resultMsg, err := RedisCli.HGetAll(context.Background(), "msg:"+msgId).Result()
		if err != nil {
			return nil, err
		} else if len(resultMsg) == 0 {
			//过期了，删掉这个目录里面的
			RedisCli.LRem(context.Background(), target+uid, 0, msgId) //不处理错误，白兰
		} else {
			msgList[itrCounter] = data_model.Message{
				Id:       resultMsg["id"],
				SendTime: resultMsg["send_time"],
				FromId:   resultMsg["from_id"],
				TargetId: resultMsg["target_id"],
				MsgType:  resultMsg["msg_type"],
				Content:  resultMsg["content"],
			}
			itrCounter++
		}
	}
	msgList = msgList[:itrCounter]

	return msgList, nil
}

func GetMsgsOfTwo(uid1 string, uid2 string) ([]data_model.Message, error) {
	//获取双方目录
	msgIdList1Send, err := RedisCli.LRange(context.Background(), "send:"+uid1, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	msgIdList2Send, err := RedisCli.LRange(context.Background(), "send:"+uid2, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	msgIdList1Recv, err := RedisCli.LRange(context.Background(), "recv:"+uid1, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	msgIdList2Recv, err := RedisCli.LRange(context.Background(), "recv:"+uid2, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	sentList1 := intersect.Simple(msgIdList1Send, msgIdList2Recv)
	sentList2 := intersect.Simple(msgIdList2Send, msgIdList1Recv)
	sumUp := append(sentList1, sentList2)
	msgIdList := make([]string, len(sumUp))
	for i, v := range sumUp {
		msgIdList[i] = fmt.Sprint(v)
	}

	//创建一个List用于存放：
	var msgList = make([]data_model.Message, len(sentList1))

	//循环获取message：
	itrCounter := 0
	for _, msgId := range msgIdList {
		resultMsg, err := RedisCli.HGetAll(context.Background(), "msg:"+msgId).Result()
		if err != nil {
			return nil, err
		} else if len(resultMsg) == 0 {
			//过期了，删掉这个目录里面的
			RedisCli.LRem(context.Background(), "recv:"+uid1, 0, msgId) //不处理错误，白兰
			RedisCli.LRem(context.Background(), "recv:"+uid2, 0, msgId)
			RedisCli.LRem(context.Background(), "send:"+uid1, 0, msgId)
			RedisCli.LRem(context.Background(), "send:"+uid2, 0, msgId)

		} else {
			msgList[itrCounter] = data_model.Message{
				Id:       resultMsg["id"],
				SendTime: resultMsg["send_time"],
				FromId:   resultMsg["from_id"],
				TargetId: resultMsg["target_id"],
				MsgType:  resultMsg["msg_type"],
				Content:  resultMsg["content"],
			}
			itrCounter++
		}
	}
	msgList = msgList[:itrCounter]

	return msgList, nil
}

func GetMsgOfGroup(gid string) ([]data_model.Message, error) {
	//获取目录
	msgIdList, err := RedisCli.LRange(context.Background(), "recv:"+gid, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	//创建一个List用于存放：
	var msgList = make([]data_model.Message, len(msgIdList))

	//循环获取message：
	itrCounter := 0
	for _, msgId := range msgIdList {
		resultMsg, err := RedisCli.HGetAll(context.Background(), "msg:"+msgId).Result()
		if err != nil {
			return nil, err
		} else if len(resultMsg) == 0 {
			//过期了，删掉这个目录里面的
			RedisCli.LRem(context.Background(), "recv:"+gid, 0, msgId)

		} else {
			msgList[itrCounter] = data_model.Message{
				Id:       resultMsg["id"],
				SendTime: resultMsg["send_time"],
				FromId:   resultMsg["from_id"],
				TargetId: resultMsg["target_id"],
				MsgType:  resultMsg["msg_type"],
				Content:  resultMsg["content"],
			}
			itrCounter++
		}
	}
	msgList = msgList[:itrCounter]

	return msgList, nil
}

func SetAuth(uid string, auth string) error {
	return RedisCli.HSet(context.Background(), "auth", uid, auth).Err()
}

func GetAuth(uid string) (string, error) {
	m := RedisCli.HGet(context.Background(), "auth", uid)
	auth := m.String()
	err := m.Err()
	return auth, err
}
