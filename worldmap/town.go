package worldmap

type Town struct {
	Name       string
	TX1        int
	TY1        int
	TX2        int
	TY2        int
	SX1        int
	SY1        int
	SX2        int
	SY2        int
	Horizontal bool
	Buildings  []Building
}
