package scanfile

type Counter struct {
	Num int
	max int
}

func (st *Counter) Add() {
	st.Num = st.Num + 1
}

func (st *Counter) IsMax() bool {
	return st.Num >= st.max
}

func (st *Counter) SetMax(max int) {
	st.max = max
}

func InitCounter(max int) *Counter {
	p := new(Counter)
	p.SetMax(max)
	return p
}
