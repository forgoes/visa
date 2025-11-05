package http

import "fmt"

var EmailCaptchaKey = func(email string) string {
	return fmt.Sprintf("/captchas/?email=%s", email)
}

var EmailTokenKey = func(id uint, email string) string {
	return fmt.Sprintf("/tokens/?id=%d&email=%s", id, email)
}
