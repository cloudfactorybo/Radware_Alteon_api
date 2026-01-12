package config

type Config struct {
	Server  ServerConfig
	Alteons []AlteonConfig
}

type ServerConfig struct {
	Host string
	Port string
}

type AlteonConfig struct {
	Name               string
	BaseURL            string
	Username           string
	Password           string
	InsecureSkipVerify bool
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "127.0.0.1",
			Port: "5687",
		},
		Alteons: []AlteonConfig{
			{
				Name:               "DELIZIA-ALTEON-01",
				BaseURL:            "https://192.168.42.110",
				Username:           "api",
				Password:           "apiDelizia4321.CLF",
				InsecureSkipVerify: true,
			},
			{
				Name:               "DELIZIA-ALTEON-02",
				BaseURL:            "https://192.168.42.111",
				Username:           "api",
				Password:           "apiDelizia4321.CLF",
				InsecureSkipVerify: true,
			},
		},
	}
}
