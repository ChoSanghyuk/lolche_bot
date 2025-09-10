package lolcheBot

type DecOptMsg struct {
	Title string
	Rcmds []string
	Ids   []int
}

type Command string

const (
	help      Command = "/help"
	mode      Command = "/mode"
	switching Command = "/switch"
	updating  Command = "/update"
	reset     Command = "/reset"
	done      Command = "/done"
	fix       Command = "/fix"
)

func allCommands() []Command {
	return []Command{
		help,
		mode,
		switching,
		updating,
		reset,
		done,
		fix,
	}
}

// type Status uint

// const (
// 	waiting Status = iota
// 	provided
// 	selected
// 	completed
// )

// type MsgTitle string

const (
	titleCompletionList   string = "완료 목록"
	titleNormalDeck       string = "일반 덱"
	titleSpecDeck         string = "증강 덱"
	titleWhetherCompleted        = "완료 여부"
)

type Mode bool

const (
	MainMode Mode = true
	PbeMode  Mode = false
)

func (m Mode) Str() string {
	if m {
		return "정규 모드"
	} else {
		return "pbe 모드"
	}
}
