package runtime

import (
	"context"
)

type Runtime struct {
	Flags      *Flags
	Config     *Config
	Redis      *Redis
	PG         *Postgres
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

	pg, err := newPostgres(r.Config)
	if err != nil {
		return nil, err
	}
	r.PG = pg

	client, err := newEtcdClient(&r.Config.Deps.Etcd)
	if err != nil {
		return nil, err
	}
	r.EtcdClient = client

	return r, nil
}

func (r *Runtime) Close(ctx context.Context) error {
	err := r.Redis.Close()
	if err != nil {
		return err
	}

	err = r.PG.Close()
	if err != nil {
		return err
	}

	err = r.EtcdClient.Close()
	if err != nil {
		return err
	}

	return nil
}
