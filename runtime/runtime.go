package runtime

import (
	"context"
)

type Runtime struct {
	Flags      *Flags
	Config     *Config
	Redis      *Redis
	EtcdClient *EtcdClient
}

func NewRuntime() (*Runtime, error) {
	r := &Runtime{}

	flags, err := parseFlags()
	if err != nil {
		return nil, err
	}
	r.Flags = flags

	config, err := loadConfig(flags.ConfigFile)
	if err != nil {
		return nil, err
	}
	r.Config = config

	if err := initLogger(&r.Config.Logging); err != nil {
		return nil, err
	}

	redis, err := newRedis(r)
	if err != nil {
		return nil, err
	}
	r.Redis = redis
	/*
		client, err := NewEtcdClient(&r.Config.Etcd)
		if err != nil {
			return nil, err
		}
		r.EtcdClient = client

	*/

	return r, nil
}

func (r *Runtime) Close(ctx context.Context) error {
	err := r.Redis.Close()
	if err != nil {
		return err
	}

	return nil
}
