package forjfile

type UsersStruct map[string]*UserStruct

func (u UsersStruct) mergeFrom(from UsersStruct) UsersStruct {
	if from == nil {
		return u
	}
	if u == nil {
		u = make(UsersStruct)
	}
	for k, userFrom := range from {
		if user, found := u[k]; found {
			user.mergeFrom(userFrom)
		} else {
			u[k] = userFrom
		}
	}
	return u
}
