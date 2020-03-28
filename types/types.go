package types

// Gen is the game generation context (i.e. pokecrystal is gen 2, whereas pokeemerald is gen 3)
type Gen int

// Generation types
const (
	GEN2 Gen = iota
	GEN3
)
