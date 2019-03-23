package worldmap

type Building struct {
	X1 int
	Y1 int
	X2 int
	Y2 int
	T  BuildingType
}

type BuildingType int

const (
	Residential BuildingType = iota
	Commercial
)
