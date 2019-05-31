package worldmap

type Town struct {
	Name       string
	TownArea   Area
	StreetArea Area
	Horizontal bool
	Farm       bool
	Buildings  []Building
}

func NewTown(name string, x1, y1, x2, y2, sX1, sY1, sX2, sY2 int, horizontal, farm bool) *Town {
	t := Town{}
	t.Name = name
	t.TownArea = Area{Coordinates{x1, y1}, Coordinates{x2, y2}}
	t.StreetArea = Area{Coordinates{sX1, sY1}, Coordinates{sX2, sY2}}
	t.Horizontal = horizontal
	t.Farm = farm

	return &t
}
