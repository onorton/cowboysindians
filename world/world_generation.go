package world

import (
	"bytes"
	"io/ioutil"
	"math"
	"math/rand"
	"strings"

	"github.com/onorton/cowboysindians/item"
	"github.com/onorton/cowboysindians/npc"
	"github.com/onorton/cowboysindians/structs"
	"github.com/onorton/cowboysindians/worldmap"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func GenerateWorld(filename string, width, height int) ([]*npc.Mount, []*npc.Enemy, []*npc.Npc) {
	world := worldmap.NewWorld(width, height)

	towns := make([]worldmap.Town, 0)
	buildings := make([]worldmap.Building, 0)

	for i := 0; i < 2; i++ {
		generateTown(world, &towns, &buildings)
	}
	generatePaths(world, towns)

	// Generate building outside town
	generateBuildingOutsideTown(world, &towns, &buildings)

	placeSignposts(world, towns)
	addItemsToBuildings(world, buildings)

	mounts := generateMounts(world, buildings, 5)
	enemies := generateEnemies(world, 2)
	npcs := generateNpcs(world, towns, buildings, 10)

	worldJson, err := world.MarshalJSON()
	check(err)
	buffer := bytes.NewBufferString("{")
	buffer.WriteString("\"World\":")
	buffer.Write(worldJson)
	buffer.WriteString("}")

	err = ioutil.WriteFile(filename, buffer.Bytes(), 0644)
	check(err)
	return mounts, enemies, npcs
}

func addItemsToBuildings(world worldmap.World, buildings []worldmap.Building) {
	for _, b := range buildings {

		// Consider inner area (exclude walls)
		x1, y1 := b.X1+1, b.Y1+1
		x2, y2 := b.X2-1, b.Y2-1

		// If Saloon, place chairs and tables
		if b.T == worldmap.Saloon {

			// Chairs and tables should not be close to counter or front door
			if b.DoorLocation.X == b.X1 {
				x1 += 1
				x2 -= 3
			} else if b.DoorLocation.X == b.X2 {
				x1 += 3
				x2 -= 1
			} else if b.DoorLocation.Y == b.Y1 {
				y1 += 1
				y2 -= 3
			} else if b.DoorLocation.Y == b.Y2 {
				y1 += 3
				y2 -= 1
			}

			for x := x1; x <= x2; x += 4 {
				for y := y1; y <= y2; y += 4 {
					// If a free 3x3 space exists, add a table and four chairs
					if x >= x1 && x+2 <= x2 && y >= y1 && y+2 <= y2 {
						world.PlaceItem(x+1, y+1, item.NewNormalItem("table"))
						world.PlaceItem(x, y+1, item.NewNormalItem("chair"))
						world.PlaceItem(x+1, y, item.NewNormalItem("chair"))
						world.PlaceItem(x+2, y+1, item.NewNormalItem("chair"))
						world.PlaceItem(x+1, y+2, item.NewNormalItem("chair"))
					}
				}
			}

		} else {

			buildingArea := (x2 - x1) * (y2 - y1)

			numOfItems := buildingArea / 2

			for i := 0; i < numOfItems; i++ {

				// Select a random location
				x := x1 + rand.Intn(x2-x1)
				y := y1 + rand.Intn(y2-y1)

				if world.IsPassable(x, y) {
					world.PlaceItem(x, y, item.GenerateItem())
				} else {
					i--
				}
			}
		}
	}

}

// Generate a rectangular building and place on map
func generateBuildingOutsideTown(world worldmap.World, towns *[]worldmap.Town, buildings *[]worldmap.Building) {
	width := world.Width()
	height := world.Height()

	validBuilding := false

	// Keeps trying until a usable building position and size found
	for !validBuilding {
		// Has to be at least three squares wide
		buildingWidth := rand.Intn(3) + 3
		buildingHeight := rand.Intn(3) + 3

		posWidth := 0
		negWidth := 0
		posHeight := 0
		negHeight := 0

		if buildingWidth%2 == 0 {
			posWidth, negWidth = buildingWidth/2, buildingWidth/2
		} else {
			posWidth, negWidth = (buildingWidth+1)/2, buildingWidth/2
		}

		if buildingHeight%2 == 0 {
			posHeight, negHeight = buildingHeight/2, buildingHeight/2
		} else {
			posHeight, negHeight = (buildingHeight+1)/2, buildingHeight/2
		}

		// stop it from reaching the edges of the map
		cX := rand.Intn(width-posWidth) + negWidth
		cY := rand.Intn(height-posHeight) + negHeight

		x1, y1 := cX-negWidth, cY-negHeight
		x2, y2 := cX+posWidth, cY+posHeight

		b := worldmap.Building{x1, y1, x2, y2, worldmap.Residential, nil}
		validBuilding = isValid(x1, y1, width, height) && isValid(x2, y2, width, height) && !overlap(*buildings, b) && !inTowns(*towns, b)

		if validBuilding {
			// Add walls
			for x := x1; x <= x2; x++ {
				world.NewTile("wall", x, y1)
				world.NewTile("wall", x, y2)
			}

			for y := y1; y <= y2; y++ {
				world.NewTile("wall", x1, y)
				world.NewTile("wall", x2, y)
			}

			// Add door randomly as long it's not a corner
			wallSelection := rand.Intn(4)

			// Wall selection is North, South, East, West
			doorX, doorY := 0, 0
			switch wallSelection {
			case 0:
				doorX = x1 + 1 + rand.Intn(buildingWidth-2)
				doorY = y1
			case 1:
				doorX = x1 + 1 + rand.Intn(buildingWidth-2)
				doorY = y2
			case 2:
				doorY = y1 + 1 + rand.Intn(buildingHeight-2)
				doorX = x2
			case 3:
				doorY = y1 + 1 + rand.Intn(buildingHeight-2)
				doorX = x1
			}
			world.NewTile("door", doorX, doorY)
			b.DoorLocation = &worldmap.Coordinates{doorX, doorY}

			// Add number of windows accbording total perimeter of building
			perimeter := 2*(y2-y1) + 2*(x2-x1)
			minNumWindows := perimeter / 5
			maxNumWindows := perimeter / 3
			numWindows := minNumWindows + rand.Intn(maxNumWindows-minNumWindows)

			for i := 0; i < numWindows; i++ {

				wallSelection = rand.Intn(4)
				wX, wY := 0, 0

				switch wallSelection {
				case 0:
					doorX := x1 + 1 + rand.Intn(buildingWidth-2)
					wX, wY = doorX, y1
				case 1:
					doorX := x1 + 1 + rand.Intn(buildingWidth-2)
					wX, wY = doorX, y2
				case 2:
					doorY := y1 + 1 + rand.Intn(buildingHeight-2)
					wX, wY = x2, doorY
				case 3:
					doorY := y1 + 1 + rand.Intn(buildingHeight-2)
					wX, wY = x1, doorY
				}

				// If a door is not in place, add window. Otherwise, try again.
				if wX != doorX && wY != doorY {
					world.NewTile("window", wX, wY)
				} else {
					i--
				}
			}

			*buildings = append(*buildings, b)
		}
	}
}

func randomBuildingType(buildings *[]worldmap.Building) worldmap.BuildingType {
	// 1/2 chance of being residential
	residential := rand.Intn(2) == 0

	if residential {
		return worldmap.Residential
	} else {
		for {
			commercialType := worldmap.BuildingType(rand.Intn(3) + 1)
			// Only one sheriff
			if commercialType == worldmap.Sheriff {

				sheriffExists := false
				for _, b := range *buildings {
					if b.T == worldmap.Sheriff {
						sheriffExists = true
					}
				}
				if !sheriffExists {
					return commercialType
				}
			} else {
				return commercialType
			}
		}
	}
}

func generateBuildingInTown(world worldmap.World, t *worldmap.Town, buildings *[]worldmap.Building) {

	validBuilding := false

	buildingType := randomBuildingType(buildings)

	// Keeps trying until a usable building position and size found
	for !validBuilding {

		// Must be at least 3 in each dimension

		// Width along the street
		buildingWidth := rand.Intn(5) + 3
		depth := rand.Intn(5) + 3
		// Commerical buildings would be larger
		if buildingType != worldmap.Residential {
			buildingWidth = rand.Intn(10) + 8
			depth = rand.Intn(10) + 8
		}

		posWidth := 0
		negWidth := 0

		if buildingWidth%2 == 0 {
			posWidth, negWidth = buildingWidth/2, buildingWidth/2
		} else {
			posWidth, negWidth = (buildingWidth+1)/2, buildingWidth/2
		}

		x1, y1 := 0, 0
		x2, y2 := 0, 0

		sideOfStreet := rand.Intn(2) == 0

		if t.Horizontal {
			centreAlongStreet := t.SX1 + 1 + rand.Intn(t.SX2-t.SX1-2)
			x1 = centreAlongStreet - negWidth
			x2 = centreAlongStreet + posWidth

			if sideOfStreet {
				y2 = t.SY1 - 1
				y1 = y2 - depth
			} else {
				y1 = t.SY2 + 1
				y2 = y1 + depth
			}
		} else {
			centreAlongStreet := t.SY1 + 1 + rand.Intn(t.SY2-t.SY1-2)
			y1 = centreAlongStreet - negWidth
			y2 = centreAlongStreet + posWidth

			if sideOfStreet {
				x2 = t.SX1 - 1
				x1 = x2 - depth
			} else {
				x1 = t.SX2 + 1
				x2 = x1 + depth
			}
		}

		b := worldmap.Building{x1, y1, x2, y2, buildingType, nil}

		validBuilding = inTown(*t, b) && !overlap(*buildings, b)

		if validBuilding {
			// Add walls
			for x := x1; x <= x2; x++ {
				world.NewTile("wall", x, y1)
				world.NewTile("wall", x, y2)
			}

			for y := y1; y <= y2; y++ {
				world.NewTile("wall", x1, y)
				world.NewTile("wall", x2, y)
			}

			doorX, doorY := 0, 0
			// Door is on side facing street
			if t.Horizontal {
				doorX = x1 + 1 + rand.Intn(buildingWidth-2)
				if sideOfStreet {
					doorY = y2
				} else {
					doorY = y1
				}
			} else {
				doorY = y1 + 1 + rand.Intn(buildingWidth-2)
				if sideOfStreet {
					doorX = x2
				} else {
					doorX = x1
				}
			}
			world.NewTile("door", doorX, doorY)
			b.DoorLocation = &worldmap.Coordinates{doorX, doorY}

			// If not residential add counter
			if b.T != worldmap.Residential {
				// Choose the side flap will appear
				flapSide := rand.Intn(2) == 0
				if t.Horizontal {
					counterY := 0
					if sideOfStreet {
						counterY = y1 + 2
					} else {
						counterY = y2 - 2
					}

					for x := x1 + 1; x < x2; x++ {
						world.NewTile("counter", x, counterY)
					}

					if flapSide {
						world.NewTile("counter flap", x1+1, counterY)
					} else {
						world.NewTile("counter flap", x2-1, counterY)
					}
				} else {
					counterX := 0
					if sideOfStreet {
						counterX = x1 + 2
					} else {
						counterX = x2 - 2
					}

					for y := y1 + 1; y < y2; y++ {
						world.NewTile("counter", counterX, y)
					}

					if flapSide {
						world.NewTile("counter flap", counterX, y1+1)
					} else {
						world.NewTile("counter flap", counterX, y2-1)
					}
				}

			}
			// Add number of windows according total perimeter of building
			perimeter := 2*buildingWidth + 2*depth
			minNumWindows := perimeter / 5
			maxNumWindows := perimeter / 3
			numWindows := minNumWindows + rand.Intn(maxNumWindows-minNumWindows)

			for i := 0; i < numWindows; i++ {

				wallSelection := rand.Intn(4)
				wX, wY := 0, 0

				switch wallSelection {
				case 0:
					windowX := x1 + 1 + rand.Intn(x2-x1-2)
					wX, wY = windowX, y1
				case 1:
					windowX := x1 + 1 + rand.Intn(x2-x1-2)
					wX, wY = windowX, y2
				case 2:
					windowY := y1 + 1 + rand.Intn(y2-y1-2)
					wX, wY = x2, windowY
				case 3:
					windowY := y1 + 1 + rand.Intn(y2-y1-2)
					wX, wY = x1, windowY
				}

				// If a door is not in place, add window. Otherwise, try again.
				if wX != doorX && wY != doorY {
					world.NewTile("window", wX, wY)
				} else {
					i--
				}

			}

			*buildings = append(*buildings, b)
			t.Buildings = append(t.Buildings, b)
		}
	}
}

// Generate small town (single street with buildings)
func generateTown(world worldmap.World, towns *[]worldmap.Town, buildings *[]worldmap.Building) {
	// Generate area of town
	width := world.Width()
	height := world.Height()

	validTown := false

	for !validTown {
		townWidth := 15 + rand.Intn(30)
		townHeight := 15 + rand.Intn(30)

		posWidth := 0
		negWidth := 0
		posHeight := 0
		negHeight := 0

		if townWidth%2 == 0 {
			posWidth, negWidth = townWidth/2, townWidth/2
		} else {
			posWidth, negWidth = (townWidth+1)/2, townWidth/2
		}

		if townHeight%2 == 0 {
			posHeight, negHeight = townHeight/2, townHeight/2
		} else {
			posHeight, negHeight = (townHeight+1)/2, townHeight/2
		}

		// stop it from reaching the edges of the map
		cX := rand.Intn(width-posWidth) + negWidth
		cY := rand.Intn(height-posHeight) + negHeight

		x1, y1 := cX-negWidth, cY-negHeight
		x2, y2 := cX+posWidth, cY+posHeight

		streetBreadth := 1 + rand.Intn(5)

		streetX1, streetY1 := 0, 0
		streetX2, streetY2 := 0, 0

		horizontalStreet := rand.Intn(2) == 0
		if horizontalStreet {
			streetX1, streetY1 = x1, cY-streetBreadth/2
			streetX2, streetY2 = x2, cY+(streetBreadth+1)/2
		} else {
			streetX1, streetY1 = cX-streetBreadth/2, y1
			streetX2, streetY2 = cX+(streetBreadth+1)/2, y2
		}

		t := &worldmap.Town{generateTownName(), x1, y1, x2, y2, streetX1, streetY1, streetX2, streetY2, horizontalStreet, make([]worldmap.Building, 0)}

		validTown := isValid(x1, y1, width, height) && isValid(x2, y2, width, height) && !townsOverlap(*towns, *t)
		if validTown {

			//Select random number of buildings, assuming 2 on each side of street
			minNumBuildings, maxNumBuildings := int(math.Max(1, float64(townWidth/10))), int(math.Max(1, float64(townWidth/5)))
			numBuildings := minNumBuildings + rand.Intn(maxNumBuildings-minNumBuildings)
			// Generate a number of buildings
			for i := 0; i < numBuildings; i++ {
				generateBuildingInTown(world, t, buildings)
			}
			*towns = append(*towns, *t)
			break
		}
	}
}

func generateTownName() string {
	noun := npc.Names.Towns["Nouns"][rand.Intn(len(npc.Names.Towns["Nouns"]))]
	withAdjective := rand.Intn(2) == 0
	if withAdjective {
		adjective := npc.Names.Towns["Adjectives"][rand.Intn(len(npc.Names.Towns["Adjectives"]))]
		name := noun

		joined := rand.Intn(2) == 0
		if joined {
			name = adjective + strings.ToLower(noun)
		} else {
			name = adjective + " " + noun
		}

		return name
	}
	return noun
}

type path struct {
	curves []func(float64) worldmap.Coordinates
	width  int
}

func generatePaths(world worldmap.World, towns []worldmap.Town) {
	// Create tiles in towns
	for _, t := range towns {
		for y := t.SY1; y <= t.SY2; y++ {
			for x := t.SX1; x <= t.SX2; x++ {
				world.NewTile("path", x, y)
			}
		}
	}

	// Determine which towns should be connected

	type connection struct {
		first  int
		second int
	}

	connections := make([]connection, 0)
	visitedTowns := structs.Initialise()
	queue := structs.Queue{}

	// choose a starting town
	queue.Enqueue(rand.Intn(len(towns)))

	for !queue.IsEmpty() {
		t := queue.Dequeue().(int)
		if !visitedTowns.Exists(t) {
			visitedTowns.Add(t)
			// Choose up to 3 towns to connect to
			num := 1 + rand.Intn(3)
			for i := 0; i < num; i++ {
				newT := rand.Intn(len(towns))
				if newT != t {
					connectionExists := false
					for _, c := range connections {
						if (c.first == t && c.second == newT) || (c.first == newT && c.second == t) {
							connectionExists = true
							break
						}
					}

					if connectionExists {
						continue
					}

					connections = append(connections, connection{t, newT})
					queue.Enqueue(newT)
				} else {
					i--
				}
			}
		}
	}

	// Define paths mathematically
	paths := make([]path, len(connections))

	for i, c := range connections {
		t1 := towns[c.first]
		t2 := towns[c.second]

		t1StreetPoints := make([]worldmap.Coordinates, 2)
		t2StreetPoints := make([]worldmap.Coordinates, 2)
		t1Width, t2Width := 0, 0

		if t1.Horizontal {
			t1StreetPoints[0] = worldmap.Coordinates{t1.SX1, (t1.SY1 + t1.SY2) / 2}
			t1StreetPoints[1] = worldmap.Coordinates{t1.SX2, (t1.SY1 + t1.SY2) / 2}
			t1Width = t1.SY2 - t1.SY1 + 1
		} else {
			t1StreetPoints[0] = worldmap.Coordinates{(t1.SX1 + t1.SX2) / 2, t1.SY1}
			t1StreetPoints[1] = worldmap.Coordinates{(t1.SX1 + t1.SX2) / 2, t1.SY2}
			t1Width = t1.SX2 - t1.SX1 + 1
		}

		if t2.Horizontal {
			t2StreetPoints[0] = worldmap.Coordinates{t2.SX1, (t2.SY1 + t2.SY2) / 2}
			t2StreetPoints[1] = worldmap.Coordinates{t2.SX2, (t2.SY1 + t2.SY2) / 2}
			t2Width = t2.SY2 - t2.SY1 + 1
		} else {
			t2StreetPoints[0] = worldmap.Coordinates{(t2.SX1 + t2.SX2) / 2, t2.SY1}
			t2StreetPoints[1] = worldmap.Coordinates{(t2.SX1 + t2.SX2) / 2, t2.SY2}
			t2Width = t2.SX2 - t2.SX1 + 1
		}

		width := (t1Width + t2Width) / 2

		intersects := true
		curves := make([]func(float64) worldmap.Coordinates, width+1)
		for intersects {
			intersects = false

			start := t1StreetPoints[rand.Intn(2)]
			end := t2StreetPoints[rand.Intn(2)]

			// Calculate all start and end points
			startPoints := make([]worldmap.Coordinates, width+1)
			endPoints := make([]worldmap.Coordinates, width+1)

			// Pick control points of the bezier curve
			// They should be close enough to the line to not have too tight corners
			r := int(worldmap.Distance(start.X, start.Y, end.X, end.Y) / 2)
			centre := worldmap.Coordinates{(start.X + end.X) / 2, (start.Y + end.Y) / 2}
			c1 := worldmap.Coordinates{centre.X + int(math.Pow(-1.0, float64(rand.Intn(2)))*float64(rand.Intn(r))),
				centre.Y + int(math.Pow(-1.0, float64(rand.Intn(2)))*float64(rand.Intn(r)))}
			c2 := worldmap.Coordinates{centre.X + int(math.Pow(-1.0, float64(rand.Intn(2)))*float64(rand.Intn(r))),
				centre.Y + int(math.Pow(-1.0, float64(rand.Intn(2)))*float64(rand.Intn(r)))}

			for j := 0; j <= width; j++ {
				if t1.Horizontal {
					startPoints[j] = worldmap.Coordinates{start.X, start.Y + j - width/2}
				} else {
					startPoints[j] = worldmap.Coordinates{start.X + j - width/2, start.Y}
				}

				if t2.Horizontal {
					endPoints[j] = worldmap.Coordinates{end.X, end.Y + j - width/2}
				} else {
					endPoints[j] = worldmap.Coordinates{end.X + j - width/2, end.Y}
				}
			}

			curve := func(start, end worldmap.Coordinates) func(t float64) worldmap.Coordinates {
				return func(t float64) worldmap.Coordinates {
					// This is just definition of a cubic bezier curve
					x := int(math.Pow(1-t, 3)*float64(start.X) + 3*(1-t)*(1-t)*t*float64(c1.X) + 3*(1-t)*t*t*float64(c2.X) + math.Pow(t, 3)*float64(end.X))
					y := int(math.Pow(1-t, 3)*float64(start.Y) + 3*(1-t)*(1-t)*t*float64(c1.Y) + 3*(1-t)*t*t*float64(c2.Y) + math.Pow(t, 3)*float64(end.Y))
					return worldmap.Coordinates{x, y}
				}
			}

			// check that path does not intersect towns
			minStep := 1.0 / (2 * math.Max(math.Abs(float64(end.X-start.X)), math.Abs(float64(end.Y-start.Y))))

			for t := 0.0; t <= 1.0; t += minStep {
				point := curve(start, end)(t)
				for _, town := range towns {
					if point.X > town.TX1 && point.X < town.TX2 && point.Y > town.TY1 && point.Y < town.TY2 {
						intersects = true
					}
				}
				if intersects {
					break
				}

			}

			if !intersects {
				for j := 0; j <= width; j++ {
					curves[j] = curve(startPoints[j], endPoints[j])
				}
			}
		}
		paths[i] = path{curves, width}
	}

	// Create tiles for paths

	for _, path := range paths {
		generatePath(world, path)
	}

}

func generatePath(world worldmap.World, path path) {
	start := path.curves[0](0.0)
	end := path.curves[0](1.0)

	minStep := 1.0 / (10 * math.Max(math.Abs(float64(end.X-start.X)), math.Abs(float64(end.Y-start.Y))))

	for _, curve := range path.curves {
		for t := 0.0; t <= 1.0; t += minStep {
			curr := curve(t)
			if curr.X >= 0 && curr.X < world.Width() && curr.Y >= 0 && curr.Y < world.Height() {
				world.NewTile("path", curr.X, curr.Y)
			}
		}
	}
}

func placeSignposts(m worldmap.World, towns []worldmap.Town) {
	for _, t := range towns {
		sX, sY := 0, 0

		if t.Horizontal {
			sY = [2]int{t.SY1 - 2, t.SY2 + 2}[rand.Intn(2)]
			sX = [2]int{t.SX1, t.SX2}[rand.Intn(2)]
		} else {
			sX = [2]int{t.SX1 - 2, t.SX2 + 2}[rand.Intn(2)]
			sY = [2]int{t.SY1, t.SY2}[rand.Intn(2)]
		}

		signpost := item.NewReadable("signpost", map[string]string{"town": t.Name})
		// Make it belong to the town so if it's stolen the character who has done it has committed a CrimeEvent
		signpost.TransferOwner(t.Name)

		m.PlaceItem(sX, sY, signpost)

	}
}

func generateMounts(m worldmap.World, buildings []worldmap.Building, n int) []*npc.Mount {
	width := m.Width()
	height := m.Height()
	mounts := make([]*npc.Mount, n)
	for i := 0; i < n; i++ {
		x := rand.Intn(width)
		y := rand.Intn(height)
		if !m.IsPassable(x, y) || m.IsOccupied(x, y) || !outside(buildings, x, y) {
			i--
			continue
		}
		mounts[i] = npc.NewMount("horse", x, y, nil)
		m.Place(mounts[i], x, y)
	}
	return mounts
}

func generateEnemies(m worldmap.World, n int) []*npc.Enemy {
	width := m.Width()
	height := m.Height()
	enemies := make([]*npc.Enemy, n)
	for i := 0; i < n; i++ {
		x := rand.Intn(width)
		y := rand.Intn(height)
		if !m.IsPassable(x, y) || m.IsOccupied(x, y) {
			i--
			continue
		}
		enemies[i] = npc.NewEnemy("bandit", x, y, nil)
		m.Place(enemies[i], x, y)

	}
	return enemies
}

func generateNpcs(m worldmap.World, towns []worldmap.Town, buildings []worldmap.Building, n int) []*npc.Npc {
	width := m.Width()
	height := m.Height()
	npcs := make([]*npc.Npc, n)

	usedBuildings := make([]worldmap.Building, 0)

	commercialBuildings := make([]worldmap.Building, 0)
	for _, b := range buildings {
		if b.T != worldmap.Residential {
			commercialBuildings = append(commercialBuildings, b)
		}
	}
	i := 0

	// Place npcs in commerical buildings first
	for ; i < n && i < len(commercialBuildings); i++ {
		b := commercialBuildings[i]
		npcs[i] = placeNpcInBuilding(m, findTown(towns, b), b)
		x, y := npcs[i].GetCoordinates()
		m.Place(npcs[i], x, y)
		usedBuildings = append(usedBuildings, b)
	}

	for ; i < n; i++ {
		x, y := 0, 0

		// 50/50 chance of being placed inside or outside a building
		insideBuilding := rand.Intn(2) == 0
		if insideBuilding {
			b := buildings[rand.Intn(len(buildings))]
			usedBefore := false
			for _, building := range usedBuildings {
				if b == building {
					usedBefore = true
				}
			}

			if usedBefore {
				i--
				continue
			}
			npcs[i] = placeNpcInBuilding(m, findTown(towns, b), b)
			usedBuildings = append(usedBuildings, b)

		} else {
			x, y = rand.Intn(width), rand.Intn(height)
			if !m.IsPassable(x, y) || m.IsOccupied(x, y) {
				i--
				continue
			}
			npcs[i] = npc.NewNpc("townsman", x, y, nil)
		}
		x, y = npcs[i].GetCoordinates()
		m.Place(npcs[i], x, y)

	}
	return npcs
}

func placeNpcInBuilding(m worldmap.World, t worldmap.Town, b worldmap.Building) *npc.Npc {
	var n *npc.Npc
	for n == nil {
		x := b.X1 + 1 + rand.Intn(b.X2-b.X1-1)
		y := b.Y1 + 1 + rand.Intn(b.Y2-b.Y1-1)

		if !m.IsPassable(x, y) || m.IsOccupied(x, y) {
			continue
		}

		switch b.T {
		case worldmap.Residential:
			n = npc.NewNpc("townsman", x, y, nil)
		case worldmap.GunShop:
			n = npc.NewShopkeeper("shopkeeper", x, y, nil, t, b)
		case worldmap.Saloon:
			n = npc.NewShopkeeper("bartender", x, y, nil, t, b)
		case worldmap.Sheriff:
			n = npc.NewShopkeeper("sheriff", x, y, nil, t, b)
		}
	}
	return n
}

func findTown(towns []worldmap.Town, b worldmap.Building) worldmap.Town {
	// Find town building is in
	for _, town := range towns {
		for _, building := range town.Buildings {
			if building == b {
				return town
			}
		}
	}
	return worldmap.Town{}
}

func isValid(x, y, width, height int) bool {
	return x >= 0 && y >= 0 && x < width && y < height
}

func outside(buildings []worldmap.Building, x, y int) bool {
	for _, b := range buildings {
		if b.Inside(x, y) {
			return false
		}
	}
	return true
}

func inTown(t worldmap.Town, b worldmap.Building) bool {
	return b.X1 >= t.TX1 && b.X1 <= t.TX2 && b.X2 >= t.TX1 && b.X2 <= t.TX2 && b.Y1 >= t.TY1 && b.Y1 <= t.TY2 && b.Y2 >= t.TY1 && b.Y2 <= t.TY2
}

func inTowns(towns []worldmap.Town, b worldmap.Building) bool {
	for _, t := range towns {
		if inTown(t, b) {
			return true
		}
	}
	return false
}

func overlap(buildings []worldmap.Building, b worldmap.Building) bool {
	for _, otherBuilding := range buildings {
		if !(b.X2 < otherBuilding.X1-1 || otherBuilding.X2 < b.X1-1 || b.Y2 < otherBuilding.Y1-1 || otherBuilding.Y2 < b.Y1-1) {
			return true
		}
	}
	return false

}

func townsOverlap(towns []worldmap.Town, t worldmap.Town) bool {
	for _, otherTown := range towns {
		if !(t.TX2 < otherTown.TX1 || otherTown.TX2 < t.TX1 || t.TY2 < otherTown.TY1 || otherTown.TY2 < t.TY1) {
			return true
		}

		t1cX, t1cY := (t.TX1+t.TX2)/2, (t.TY1+t.TY2)/2
		t2cX, t2cY := (otherTown.TX1+otherTown.TX2)/2, (otherTown.TY1+otherTown.TY2)/2
		distance := math.Sqrt(math.Pow(float64(t1cX-t2cX), 2) + math.Pow(float64(t1cY-t2cY), 2))

		if distance < 40 {
			return true
		}
	}
	return false

}
