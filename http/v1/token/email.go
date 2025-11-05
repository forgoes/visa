package token

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type createEmailTokenParams struct {
	Email   string `json:"email" binding:"required"`
	Captcha string `json:"captcha" binding:"required"`
}

func verifyEmail(rt *runtime.Runtime, email, captcha string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	value, err := rt.Redis.Cli.Get(ctx, api.EmailCaptchaKey(email)).Result()
	cancel()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		} else {
			return false, err
		}
	}

	if value != captcha {
		return false, nil
	}

	return true, nil
}

func createUser(rt *runtime.Runtime, email string) (*model.User, error) {
	user := &model.User{
		Email: email,
		State: model.ENABLED.String(),
	}
	res := rt.Postgres.Create(user)

	if res.Error != nil {
		return nil, res.Error
	}

	return user, nil
}

func createToken(rt *runtime.Runtime, id uint, email string) (string, error) {
	token, err := lib.RandToken(16)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err = rt.Redis.Cli.Set(ctx, api.EmailTokenKey(id, email), token, 2*time.Minute).Err()
	cancel()
	if err != nil {
		return "", err
	}

	return token, nil
}

func CreateEmailToken(rt *runtime.Runtime, c *gin.Context) (interface{}, *api.Error) {
	var params createEmailTokenParams
	if err := c.ShouldBind(&params); err != nil {
		return nil, api.InvalidArgument(nil, err.Error())
	}

	pass, err := verifyEmail(rt, params.Email, params.Captcha)
	if err != nil {
		return nil, api.InternalServerError(err.Error())
	}
	if !pass {
		return nil, api.InvalidArgument(nil, "wrong captcha")
	}

	user := &model.User{}
	err = rt.Postgres.DB.Where("email = ?", params.Email).First(user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			user, err = createUser(rt, params.Email)
			if err != nil {
				return nil, api.InternalServerError(err.Error())
			}
			token, err := createToken(rt, user.ID, params.Email)
			if err != nil {
				return nil, api.InternalServerError(err.Error())
			}
			return map[string]interface{}{
				"pending": true,
				"user":    user,
				"token":   token,
			}, nil
		} else {
			return nil, api.InternalServerError(err.Error())
		}
	} else {
		if len(user.Password) <= 0 {
			token, err := createToken(rt, user.ID, params.Email)
			if err != nil {
				return nil, api.InternalServerError(err.Error())
			}
			return map[string]interface{}{
				"pending": true,
				"user":    user,
				"token":   token,
			}, nil
		}
	}

	return map[string]interface{}{
		"pending": false,
		"user":    user,
	}, nil
}
