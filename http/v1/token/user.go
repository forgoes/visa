package token

import (
	"errors"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type createUserTokenParams struct {
	ID       uint   `json:"id" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type profile struct {
	model.GoogleAccount
}

func createJwtToken(rt *runtime.Runtime, uid uint) (string, error) {
	var profiles []profile

	err := rt.Postgres.Model(&model.GoogleAccount{}).Joins(
		`left join "user" on "google_account"."uid" = "user"."id"`,
	).Scan(&profiles).Error
	if err != nil {
		return "", err
	}

	if len(profiles) <= 0 {
		return "", errors.New("not found")
	}

	expire := time.Hour * time.Duration(2)
	claims := lib.JWTClaims{
		Id:        profiles[0].UID,
		Gid:       profiles[0].GID,
		Email:     profiles[0].Email,
		Picture:   profiles[0].Picture,
		FirstName: profiles[0].FirstName,
		LastName:  profiles[0].FamilyName,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expire).Unix(),
			Id:        strconv.Itoa(int(profiles[0].UID)),
			IssuedAt:  0,
			Issuer:    "display.group.com",
			Subject:   "user",
			Audience:  "",
			NotBefore: 0,
		},
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := at.SignedString([]byte("todo-change"))
	if err != nil {
		return "", err
	}

	return token, nil
}

func CreateUserToken(rt *runtime.Runtime, c *gin.Context) (interface{}, *api.Error) {
	var params createUserTokenParams
	if err := c.ShouldBind(&params); err != nil {
		return nil, api.InvalidArgument(nil, err.Error())
	}

	user := &model.User{}
	err := rt.Postgres.DB.Where("id = ? AND email = ?", params.ID, params.Email).First(user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, api.InvalidArgument(nil, "user not found")
		} else {
			return nil, api.InternalServerError(err.Error())
		}
	}

	if params.Password != user.Password {
		return nil, api.InvalidArgument(nil, "invalid user or password")
	}

	token, err := createJwtToken(rt, user.ID)
	if err != nil {
		return nil, api.InternalServerError(err.Error())
	}

	return map[string]interface{}{
		"user":  user,
		"token": token,
	}, nil
}
