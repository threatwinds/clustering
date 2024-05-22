package helpers

import (
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/threatwinds/logger"
)

type cfg struct {
	ClusterPort int32
	SeedNodes   []string
	DataCenter  int32
	Rack        int32
}

var config = new(cfg)
var configOnce sync.Once

func getEnvStr(name, def string, required bool) (string, *logger.Error) {
	val := os.Getenv(name)

	if val == "" {
		if required {
			return "", Logger.ErrorF("configuration required: %s", name)
		} else {
			return def, nil
		}
	}

	return val, nil
}

func getEnvStrSlice(name, def string, required bool) ([]string, *logger.Error) {
	val, e := getEnvStr(name, def, required)
	if e != nil {
		return []string{}, e
	}

	lst := strings.Split(strings.TrimSpace(val), ",")

	for i, v := range lst {
		lst[i] = strings.TrimSpace(v)
	}

	return lst, nil
}

func getEnvInt(name string, def string, required bool) (int32, *logger.Error) {
	str, e := getEnvStr(name, def, required)
	if e != nil {
		return 0, e
	}

	val, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return 0, Logger.ErrorF(err.Error())
	}

	return int32(val), nil
}

func GetCfg() *cfg {
	var e *logger.Error
	configOnce.Do(func() {
		config.ClusterPort, e = getEnvInt("CLUSTER_PORT", "10731", false)
		if e != nil {
			os.Exit(1)
		}

		config.SeedNodes, e = getEnvStrSlice("CLUSTER_SEED_NODES", "", true)
		if e != nil {
			os.Exit(1)
		}

		config.DataCenter, e = getEnvInt("NODE_DATA_CENTER", "1", false)
		if e != nil {
			os.Exit(1)
		}

		config.Rack, e = getEnvInt("NODE_RACK", "1", false)
		if e != nil {
			os.Exit(1)
		}
	})

	return config
}
