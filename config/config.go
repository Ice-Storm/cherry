package config

import (
	"cherrychain/clogging"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	protocol "github.com/libp2p/go-libp2p-protocol"
	"github.com/spf13/viper"
)

var confLogger = clogging.MustGetLogger("Config")

type CherryConfig struct {
	Conf           *viper.Viper
	BootstrapPeers []string
	NetworkID      string
	ProtocolID     protocol.ID
}

// Load config file
func Load(fileName string) (*CherryConfig, error) {
	conf := viper.New()
	conf.SetConfigName(fileName)
	currentPath, err := GetCurrentPath()

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

	protocolID := conf.GetString("protocolID")

	if protocolID == "" {
		confLogger.Fatal("protocolID id not provided")
	}

	return &CherryConfig{
		Conf:           conf,
		BootstrapPeers: conf.GetStringSlice("BootstrapPeers"),
		NetworkID:      networkID,
		ProtocolID:     protocol.ID(protocolID),
	}, nil
}

func GetCurrentPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}
	if i < 0 {
		return "", errors.New(`error: Can't find "/" or "\".`)
	}
	return string(path[0 : i+1]), nil
}
