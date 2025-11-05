package token

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Form struct {
	ClientId   string `form:"client_id" binding:"required"`
	Credential string `form:"credential" binding:"required"`
	SelectBy   string `form:"select_by" binding:"required"`
	GCsrfToken string `form:"g_csrf_token" binding:"required"`
}

type GoogleAccount struct {
	Iss           string `json:"iss"`
	Nbf           string `json:"nbf"`
	Aud           string `json:"aud"`
	Sub           string `json:"sub"`
	Hd            string `json:"hd"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Azp           string `json:"azp"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Iat           string `json:"iat"`
	Exp           string `json:"exp"`
	Jti           string `json:"jti"`
	Alg           string `json:"alg"`
	Kid           string `json:"kid"`
	Typ           string `json:"typ"`
}

func verifyOnline(token string) (*GoogleAccount, error) {
	resp, err := http.Get(fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", token))
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var account GoogleAccount
	err = json.Unmarshal(body, &account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func CreateGoogleToken(rt *runtime.Runtime, c *gin.Context) (interface{}, *api.Error) {
	// TODO https://developers.google.com/identity/gsi/web/guides/verify-google-id-token
	// TODO https://developers.google.com/identity/sign-in/web/backend-auth
	// CSRF https://tech.meituan.com/2018/10/11/fe-security-csrf.html
	var form Form
	if err := c.ShouldBind(&form); err != nil {
		return nil, api.InvalidArgument(nil, err.Error())
	}

	payload, err := verifyOnline(form.Credential)
	if err != nil {
		return nil, api.InternalServerError(err.Error())
	}

	account := &model.GoogleAccount{}
	err = rt.Postgres.DB.Where("gid = ?", payload.Sub).First(account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err = rt.Postgres.Transaction(func(tx *gorm.DB) error {
				// do some database operations in the transaction (use 'tx' from this point, not 'db')
				user := &model.User{
					Email: payload.Email,
					State: model.ENABLED.String(),
				}
				if e := tx.Create(user).Error; e != nil {
					// return any error will roll back
					return e
				}

				account = &model.GoogleAccount{
					GID:        payload.Sub,
					UID:        user.ID,
					Name:       payload.Name,
					Email:      payload.Email,
					FirstName:  payload.GivenName,
					FamilyName: payload.FamilyName,
					Picture:    payload.Picture,
				}

				if e := tx.Create(account).Error; e != nil {
					return e
				}

				// return nil will commit the whole transaction
				return nil
			})
			if err != nil {
				return nil, api.InternalServerError(err.Error())
			}
		} else {
			return nil, api.InternalServerError(err.Error())
		}
	}

	token, err := createJwtToken(rt, account.UID)
	if err != nil {
		return nil, api.InternalServerError(err.Error())
	}

	return fmt.Sprintf("%s?token=%s", rt.Config.Google.Redirect, token), api.Redirect(http.StatusSeeOther)
}

/*
func getGooglePublicKey(keyID string) (string, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v1/certs")
	if err != nil {
		return "", err
	}
	dat, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	myResp := map[string]string{}
	err = json.Unmarshal(dat, &myResp)
	if err != nil {
		return "", err
	}
	key, ok := myResp[keyID]
	if !ok {
		return "", errors.New("key not found")
	}
	return key, nil
}

type GoogleClaims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	FirstName     string `json:"given_name"`
	LastName      string `json:"family_name"`
	jwt.StandardClaims
}

func ValidateGoogleJWT(tokenString string) (GoogleClaims, error) {
	claimsStruct := GoogleClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) {
			pem, err := getGooglePublicKey(fmt.Sprintf("%s", token.Header["kid"]))
			if err != nil {
				return nil, err
			}
			key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pem))
			if err != nil {
				return nil, err
			}
			return key, nil
		},
	)
	if err != nil {
		return GoogleClaims{}, err
	}

	claims, ok := token.Claims.(*GoogleClaims)
	if !ok {
		return GoogleClaims{}, errors.New("Invalid Google JWT")
	}

	if claims.Issuer != "accounts.google.com" && claims.Issuer != "https://accounts.google.com" {
		return GoogleClaims{}, errors.New("iss is invalid")
	}

	if claims.Audience != "YOUR_CLIENT_ID_HERE" {
		return GoogleClaims{}, errors.New("aud is invalid")
	}

	if claims.ExpiresAt < time.Now().UTC().Unix() {
		return GoogleClaims{}, errors.New("JWT is expired")
	}

	return *claims, nil
}

	cookie, err := c.Cookie("g_csrf_token")
	if err != nil {
		return nil, api.InvalidArgument(nil, err.Error())
	}

	if len(form.GCsrfToken) <= 0 {
		return nil, api.InvalidArgument(nil, "invalid g_csrf_token")
	}

	if cookie != form.GCsrfToken {
		return nil, api.InvalidArgument(nil, "failed to verify double submit cookie")
	}
*/
