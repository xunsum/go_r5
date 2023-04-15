package consts

type AccessModeStruct struct {
	USER_MODE  int
	ADMIN_MODE int
}

func ACCESS_MODE() AccessModeStruct {
	return AccessModeStruct{
		USER_MODE:  0,
		ADMIN_MODE: 1,
	}
}
