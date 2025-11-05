package runtime

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	Cli *redis.Client
}

func newRedis(rt *Runtime) (*Redis, error) {
	cli := redis.NewClient(&redis.Options{
		Addr:         rt.Config.Deps.Redis.Address,
		Username:     rt.Config.Deps.Redis.User,
		Password:     rt.Config.Deps.Redis.Password,
		DB:           rt.Config.Deps.Redis.DB,
		MaxRetries:   rt.Config.Deps.Redis.MaxRetries,
		PoolSize:     rt.Config.Deps.Redis.PoolSize,
		MinIdleConns: rt.Config.Deps.Redis.MinIdle,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := cli.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return &Redis{Cli: cli}, nil
}

func (r *Redis) Close() error {
	return r.Cli.Close()
}
