package forjfile

// Model used by template to secure data.

type AppModel struct {
	app *AppStruct
}

func (r AppModel)Get(field string) (_ string) {
	if r.app == nil {
		return
	}
	v, _, _ := r.app.Get(field)
	return v.GetString()
}
