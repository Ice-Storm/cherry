package config

import (
	"cherrychain/common/clogging"
	"cherrychain/p2p/util"

	"github.com/spf13/viper"
)

var confLogger = clogging.MustGetLogger("Config")

type CherryConfig struct {
	Conf           *viper.Viper
	BootstrapPeers []string
	NetworkID      string
}

// Load config file
func Load(fileName string) (*CherryConfig, error) {
	conf := viper.New()
	conf.SetConfigName(fileName)
	currentPath, err := util.GetCurrentPath()

	if err != nil {
		confLogger.Fatal("Cant't get config file: ", err)
	}

	conf.AddConfigPath(currentPath)

	if err := conf.ReadInConfig(); err != nil {
		confLogger.Fatal("Fatal error config file: ", err)
	}

	networkID := conf.GetString("networkID")

	if networkID == "" {
		confLogger.Fatal("Network id not provided")
	}

	return &CherryConfig{
		Conf:           conf,
		BootstrapPeers: conf.GetStringSlice("BootstrapPeers"),
		NetworkID:      networkID,
	}, nil
}
