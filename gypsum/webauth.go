package gypsum

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/gin-gonic/gin"
)

type loginValidatorType struct {
	Cookie string
}

var loginValidator loginValidatorType

const loginCookieName = "gypsum-login-1"

func (v *loginValidatorType) check(s string) bool {
	return v.Cookie != "" && v.Cookie == s
}

func (v *loginValidatorType) login(passwordEncrypted string) (string, bool) {
	if Config.Password == passwordEncrypted {
		newCookieBytes := make([]byte, 24)
		if _, err := rand.Read(newCookieBytes); err != nil {
			panic(err)
		}
		newCookie := base64.URLEncoding.EncodeToString(newCookieBytes)
		v.Cookie = newCookie
		return newCookie, true
	} else {
		return "", false
	}
}

func authMiddleware(c *gin.Context) {
	loginCookie, _ := c.Cookie(loginCookieName)
	if loginValidator.check(loginCookie) {
		c.Next()
	} else {
		c.Data(401, "text/plain", []byte("401: Unauthorized"))
		c.Abort()
	}
}

type loginRequest struct {
	Password string `json:"password"`
}

func loginHandler(c *gin.Context) {
	req := loginRequest{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code":    8,
			"message": "provide password",
		})
		return
	}
	cookie, ok := loginValidator.login(req.Password)
	if !ok {
		c.JSON(401, gin.H{
			"code":    9,
			"message": "wrong password",
		})
		return
	}
	c.SetCookie(loginCookieName, cookie, 60*60*24*365, "/api/v1", "", false, true)
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
	})
	return
}

func getGypsumInformation(c *gin.Context) {
	loginCookie, _ := c.Cookie(loginCookieName)
	c.JSON(200, gin.H{
		"version":       BuildVersion,
		"commit":        BuildCommit,
		"password_salt": Config.PasswordSalt,
		"logged_in":     loginValidator.check(loginCookie),
	})
}
