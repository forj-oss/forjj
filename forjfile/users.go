package forjfile

type UsersStruct map[string]*UserStruct

func (u UsersStruct) mergeFrom(source string, from UsersStruct) {
	for k, userFrom := range from {
		if user, found := u[k]; found {
			user.mergeFrom(source, userFrom)
		} else {
			u[k] = userFrom
		}
	}
}
