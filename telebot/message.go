package telebot

// 2. 덱 추천은 이거야
type DecMsg struct {
	Title []string
	Rcmds [][]string
	Ids   [][]int
}

type DecResp struct {
	Type string
	Id   int
}
