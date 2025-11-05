package password

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type editParams struct {
	Token            string `json:"token" binding:"required"`
	ID               uint   `json:"id" binding:"required"`
	Email            string `json:"email" binding:"required"`
	Password         string `json:"password" binding:"required"`
	PasswordRepeated string `json:"password_repeated" binding:"required"`
}

func Edit(rt *runtime.Runtime, c *gin.Context) (interface{}, *api.Error) {
	var params editParams
	if err := c.ShouldBind(&params); err != nil {
		return nil, api.InvalidArgument(nil, err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	value, err := rt.Redis.Cli.Get(ctx, api.EmailTokenKey(params.ID, params.Email)).Result()
	cancel()
	if err != nil {
		if err == redis.Nil {
			return nil, api.InvalidArgument(nil, "no token found")
		} else {
			return nil, api.InternalServerError(err.Error())
		}
	}
	if value != params.Token {
		return nil, api.InvalidArgument(nil, "invalid token")
	}

	if params.Password != params.PasswordRepeated {
		return nil, api.InvalidArgument(nil, "not the same password")
	}

	res := rt.Postgres.Model(&model.User{}).Where(
		"id = ? AND email = ?", params.ID, params.Email,
	).Update("password", params.Password)
	if res.Error != nil {
		return nil, api.InternalServerError(res.Error.Error())
	}
	if res.RowsAffected <= 0 {
		return nil, api.InvalidArgument(nil, "record not found")
	}

	return nil, nil
}
