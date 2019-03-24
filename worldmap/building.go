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
	GunShop
	Saloon
)

func (b Building) Inside(x, y int) bool {
	return x >= b.X1 && x <= b.X2 && y >= b.Y1 && y <= b.Y2
}
