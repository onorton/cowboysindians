package worldmap

import (
	"encoding/json"
	"io/ioutil"
)

type worldConfig struct {
	Width  int
	Height int
}

var WorldConf = fetchWorldConfig()

func fetchWorldConfig() worldConfig {
	data, err := ioutil.ReadFile("data/world.json")
	check(err)
	var wc worldConfig
	err = json.Unmarshal(data, &wc)
	check(err)
	return wc
}
