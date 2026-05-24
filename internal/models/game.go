package models

// GameSlug is the URL-safe identifier for a game.
type GameSlug string

const (
	GameDualNBack GameSlug = "dual-n-back"
	GameReversi   GameSlug = "reversi"
)

// Game holds metadata for a game entry.
type Game struct {
	Slug        GameSlug
	Name        string
	Description string
	Icon        string // path under /static/images/
}

// AllGames is the registry of available games.
var AllGames = []Game{
	{
		Slug:        GameDualNBack,
		Name:        "Dual N-Back",
		Description: "A working-memory training exercise that improves fluid intelligence.",
		Icon:        "/static/images/dual-n-back.svg",
	},
	{
		Slug:        GameReversi,
		Name:        "Reversi",
		Description: "A classic two-player strategy board game, also known as Othello.",
		Icon:        "/static/images/reversi.svg",
	},
}

// FindGame returns the game matching the given slug, or false if not found.
func FindGame(slug GameSlug) (Game, bool) {
	for _, g := range AllGames {
		if g.Slug == slug {
			return g, true
		}
	}
	return Game{}, false
}
