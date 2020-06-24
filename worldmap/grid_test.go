package worldmap

import "testing"

func init() {
	terrainDataPath = "../" + terrainDataPath
}

func TestNewGridStartsWithAllGroundTilesAndCorrectSize(t *testing.T) {
	width, height := 5, 6
	grid := NewGrid(width, height)

	if grid.height() != height {
		t.Errorf("Expected height to be %d but was %d", height, grid.height())
	}

	if grid.width() != width {
		t.Errorf("Expected height to be %d but was %d", width, grid.width())
	}

	for y, row := range grid.passable {
		for x, passable := range row {
			if !passable {
				t.Errorf("Expected tile at %d, %d to be passable but it wasn't", x, y)
			}
		}
	}

	for y, row := range grid.blocksVision {
		for x, blocksVision := range row {
			if blocksVision {
				t.Errorf("Expected tile at %d, %d to not block vision but it did", x, y)
			}
		}
	}

	for y, row := range grid.door {
		for x, door := range row {
			if door != nil {
				t.Errorf("Expected tile at %d, %d to not have a door but it did", x, y)
			}
		}
	}

	for y, row := range grid.blocksVision {
		for x, blocksVision := range row {
			if blocksVision {
				t.Errorf("Expected tile at %d, %d to not block vision but it did", x, y)
			}
		}
	}

	for y, row := range grid.items {
		for x, items := range row {
			if len(items) > 0 {
				t.Errorf("Expected tile at %d, %d to not have items but it did", x, y)
			}
		}
	}
}

func TestNewTileAddsDifferentTileToLocation(t *testing.T) {
	width, height := 5, 6
	grid := NewGrid(width, height)

	type tileTestPair struct {
		tileType                        string
		passable, blocksVision, hasDoor bool
	}

	testCases := []tileTestPair{
		{"wall", false, true, false},
		{"window", false, false, false},
		{"door", false, true, true},
		{"counter flap", false, false, true},
		{"counter", false, false, false},
		{"path", true, false, false},
		{"ground", true, false, false},
	}

	for _, testCase := range testCases {
		grid.newTile(testCase.tileType, 1, 1)
		if grid.passable[1][1] != testCase.passable {
			t.Errorf("Expected passability of tile %s to be %t but was %t", testCase.tileType, testCase.passable, grid.passable[1][1])
		}

		if grid.blocksVision[1][1] != testCase.blocksVision {
			t.Errorf("Expected blocks vision of tile %s to be %t but was %t", testCase.tileType, testCase.blocksVision, grid.blocksVision[1][1])
		}

		hasDoor := grid.door[1][1] != nil
		if hasDoor != testCase.hasDoor {
			t.Errorf("Expected door value of tile %s to be %t but was %t", testCase.tileType, testCase.hasDoor, hasDoor)
		}
	}

}
