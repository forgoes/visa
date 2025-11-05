package captcha

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type createParams struct {
	Email string `json:"email" binding:"required"`
}

func sendCaptcha(rt *runtime.Runtime, to string) error {
	captcha, err := lib.GenerateCaptcha(6)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err = rt.Redis.Cli.Set(ctx, api.EmailCaptchaKey(to), captcha, 2*time.Minute).Err()
	cancel()
	if err != nil {
		return err
	}

	return rt.Email.Send(captcha, to)
}

func Create(rt *runtime.Runtime, c *gin.Context) (interface{}, *api.Error) {
	var params createParams
	if err := c.ShouldBind(&params); err != nil {
		return nil, api.InvalidArgument(nil, err.Error())
	}

	var user model.User
	err := rt.Postgres.DB.Where("email = ?", params.Email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err = sendCaptcha(rt, params.Email)
			if err != nil {
				return nil, api.InternalServerError(err.Error())
			}
			return map[string]interface{}{
				"captcha": true,
				"user":    nil,
			}, nil
		} else {
			return nil, api.InternalServerError(err.Error())
		}
	}

	if len(user.Password) <= 0 {
		err = sendCaptcha(rt, params.Email)
		if err != nil {
			return nil, api.InternalServerError(err.Error())
		}
		return map[string]interface{}{
			"captcha": true,
			"user":    user,
		}, nil
	}

	return map[string]interface{}{
		"captcha": false,
		"user":    user,
	}, nil
}
