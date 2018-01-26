package utils

type MLoop struct {
	max []int
	elemMax int
	cur []int
	eol bool
}

func NewMLoop(max ...int) (ml *MLoop) {
	ml = new(MLoop)
	ml.elemMax = len(max)
	ml.cur = make([]int, ml.elemMax)
	ml.max = max
	return
}

func (ml *MLoop)Cur() (_ []int) {
	if ml == nil {
		return
	}
	return ml.cur
}

func (ml *MLoop)Increment() (_ []int) {
	if ml == nil {
		return
	}

	ml.increment(ml.elemMax-1)

	return ml.cur
}

func (ml *MLoop)increment(index int) {
	pos := ml.cur[index]
	pos++
	if pos >= ml.max[index]{
		ml.cur[index] = 0
		if index == 0 {
			ml.eol = true
			return
		}

		ml.increment(index - 1)
	} else {
		ml.cur[index] = pos
	}
	return
}

func (ml *MLoop)Eol() bool {
	return ml.eol
}
