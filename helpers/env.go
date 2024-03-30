package helpers

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
)

type cfg struct {
	ClusterPort int
	SharedKey   string
}

var config = new(cfg)
var configOnce sync.Once

func getEnvStr(name, def string, required bool) (string, error) {
	val := os.Getenv(name)

	if val == "" {
		if required {
			return "", fmt.Errorf("configuration required: %s", name)
		} else {
			return def, nil
		}
	}

	return val, nil
}

func getEnvInt(name string, def string, required bool) (int, error) {
	str, err := getEnvStr(name, def, required)
	if err != nil {
		return 0, err
	}

	val, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}

	return val, nil
}

func GetCfg() *cfg {
	var err error
	configOnce.Do(func() {
		config.ClusterPort, err = getEnvInt("CLUSTER_PORT", "10731", false)
		if err != nil {
			log.Fatalln(err)
		}

		config.SharedKey, err = getEnvStr("CLUSTER_SHARED_KEY", "", true)
		if err != nil {
			log.Fatalln(err)
		}
	})

	return config
}
