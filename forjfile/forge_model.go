package forjfile

import "github.com/forj-oss/forjj-modules/trace"

type ForgeModel struct {
	forge *Forge
}

func(f ForgeModel)Get(object, instance, key string) (ret string) {
	if f.forge == nil {
		return
	}
	ret , _ = f.forge.GetString(object, instance, key)
	return
}

func (f ForgeModel)HasApps(rules ...string) (_ bool) {
	if f.forge == nil {
		return
	}
	if v, err := f.forge.Forjfile().HasApps(rules...) ; err != nil {
		gotrace.Error("%s", err)
	} else {
		return v
	}
	return
}
