package worldmap

type Building struct {
	Area         Area
	T            BuildingType
	DoorLocation *Coordinates
}

type BuildingType int

const (
	Residential BuildingType = iota
	GunShop
	Saloon
	Sheriff
)

func (t BuildingType) String() string {
	return [...]string{"Residential", "GunShop", "Saloon", "Sheriff"}[t]
}

func NewBuilding(x1, y1, x2, y2 int, t BuildingType) Building {
	b := Building{}
	b.Area = Area{Coordinates{x1, y1}, Coordinates{x2, y2}}
	b.T = t
	return b
}
func (b Building) Inside(x, y int) bool {
	return x >= b.Area.X1() && x <= b.Area.X2() && y >= b.Area.Y1() && y <= b.Area.Y2()
}

type Area struct {
	Start Coordinates
	End   Coordinates
}

func (a Area) X1() int {
	return a.Start.X
}

func (a Area) X2() int {
	return a.End.X
}

func (a Area) Y1() int {
	return a.Start.Y
}

func (a Area) Y2() int {
	return a.End.Y
}
