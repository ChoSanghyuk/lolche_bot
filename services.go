package lolcheBot

type Stoage interface {
	Save(mode Mode, name string) error
	// SaveMain(name string) error
	// SavePbe(name string) error
	DeleteAll(mode Mode) error
	// DeleteAllMain() error
	// DeleteAllPbe() error
	DeleteByName(mode Mode, name string) error
	// DeleteMainByName(name string) error
	// DeletePbeByName(name string) error
	All(mode Mode) ([]string, error)
	// AllMain() ([]string, error)
	// AllPbe() ([]string, error)
	Mode() Mode
	SaveMode(mode Mode)
}

type DeckCrawler interface {
	Meta(mode Mode) (dec []string, err error)
	DeckUrl(mode Mode, id string) (string, error)
	UpdateCssPath(target string) error
}
