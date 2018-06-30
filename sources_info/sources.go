package sourcesinfo

type Sources struct {
	keys map[string]string
}

func newSources() (ret *Sources) {
	ret = new(Sources)
	ret.keys = make(map[string]string)
	return
}

func (s *Sources)Set(source, key, value string) (ret *Sources) {
	if s == nil {
		ret = newSources()
	} else {
		ret = s
	}

	if value == "" {
		delete(ret.keys, key)
	} else {
		ret.keys[key] = source
	}
	return
}

func (s *Sources)Get(key string) (source string) {
	if s == nil {
		return
	}
	if v, found := s.keys[key] ; found {
		return v
	}
	return
}