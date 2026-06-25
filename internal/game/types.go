package game

type LetterStatus int

const (
	StatusCorrect LetterStatus = iota
	StatusPresent
	StatusAbsent
)

type LetterResult struct {
	Letter rune
	Status LetterStatus
}

type GameMode string

const (
	ModeDaily GameMode = "daily"
	ModeSolo  GameMode = "solo"
)

type Game struct {
	ID         string
	Target     string
	Attempts   []string
	MaxTries   int
	Mode       GameMode
	Won        bool
	GameOver   bool
	WordLength int
}
