package config

// Gen is the game generation context (i.e. pokecrystal is gen 2, whereas pokeemerald is gen 3)
type Gen int

// Generation types
const (
	GEN2 Gen = iota
	GEN3
)

// SupportsDollarSignHexNotation tells whether or not the given gen supports hexadecimal notation
// using the dollar sign prefix. (e.g. $4abc)
func SupportsDollarSignHexNotation(gen Gen) bool {
	switch gen {
	case GEN2:
		return true
	case GEN3:
		return false
	default:
		return true
	}
}

// Supports0xHexNotation tells whether or not the given gen supports hexadecimal notation
// using the 0x prefix. (e.g. 0x4abc)
func Supports0xHexNotation(gen Gen) bool {
	switch gen {
	case GEN2:
		return false
	case GEN3:
		return true
	default:
		return true
	}
}
