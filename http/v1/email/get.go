package email

import (
	"github.com/gin-gonic/gin"
)

type listParams struct {
	Offset int    `form:"offset,default=0"`
	Limit  int    `form:"limit,default=10"`
	Email  string `form:"email"`
}

func List(rt *runtime.Runtime, c *gin.Context) (interface{}, *api.Error) {
	var params listParams
	if c.Bind(&params) != nil {
		return nil, api.InvalidArgument(nil)
	}

	var user []model.User
	q := rt.Postgres.DB

	if len(params.Email) > 0 {
		q = q.Where("email = ?", params.Email)
	}

	err := q.Offset(params.Offset).Limit(params.Limit).Find(&user).Error
	if err != nil {
		return nil, api.InternalServerError()
	}

	return user, nil
}
