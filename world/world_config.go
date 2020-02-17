package world

import (
	"encoding/json"
	"io/ioutil"
)

type worldConfig struct {
	Width        int
	Height       int
	Towns        int
	Farms        int
	OutBuildings int
	Mounts       int
	Enemies      int
	Npcs         int
}

var worldConf = fetchWorldConfig()

func fetchWorldConfig() worldConfig {
	data, err := ioutil.ReadFile("data/world.json")
	check(err)
	var wc worldConfig
	err = json.Unmarshal(data, &wc)
	check(err)
	return wc
}
