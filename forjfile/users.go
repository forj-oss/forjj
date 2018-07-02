package forjfile

type UsersStruct map[string]*UserStruct

func (u UsersStruct) mergeFrom(from UsersStruct) {
	for k, userFrom := range from {
		if user, found := u[k]; found {
			user.mergeFrom(userFrom)
		} else {
			u[k] = userFrom
		}
	}
}
