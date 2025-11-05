package runtime

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Logging struct {
	Level        string
	EnableCaller string
	EnableStack  string
	Handler      string
}

type App struct {
	StopTimeout int
}

type GRPC struct {
	Host                 string
	Port                 int
	MaxConcurrentStreams uint32
	MaxTcpConnections    int
}

type Gateway struct {
	Host     string
	Port     int
	Upstream string
}

type HTTP struct {
	Debug bool
	Host  string
	Port  int
}

type EtcdConfig struct {
	Endpoints    []string
	DialTimeout  int
	ReadTimeout  int
	WriteTimeout int
}

type PostgresConfig struct {
	Host     string
	User     string
	Password string
	DB       string
	Port     int
	SSL      bool
	TimeZone string
	Pool     struct {
		MaxIdle int
		MaxOpen int
		MaxLife int
	}
}

type RedisConfig struct {
	Address    string
	User       string
	Password   string
	DB         int
	MaxRetries int
	PoolSize   int
	MinIdle    int
}

type EmailConfig struct {
	Host     string
	Port     int
	From     string
	Password string
	Template string
}

type GoogleConfig struct {
	Redirect string
}

type Config struct {
	App     App
	Logging Logging
	Serve   struct {
		GRPC    GRPC
		Gateway Gateway
		HTTP    HTTP
	}
	Deps struct {
		Redis    RedisConfig
		Etcd     EtcdConfig
		Postgres PostgresConfig
		Email    EmailConfig
	}
}

func loadConfig(configFile string) (*Config, error) {
	config := &Config{}

	content, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	err = toml.Unmarshal(content, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
