package commands

import "flag"

type commands struct {
	Port  int
	Dest  string
	Fconf string
}

func CommandInit() *commands {
	port := flag.Int("sp", 3000, "listen port")
	dest := flag.String("d", "", "Dest MultiAddr String")
	configFile := flag.String("f", "cherry", "Config file name")
	flag.Parse()
	return &commands{
		Port:  *port,
		Dest:  *dest,
		Fconf: *configFile,
	}
}
