package gypsum

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"runtime"

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
		//newCookieBytes := make([]byte, 24)
		//if _, err := rand.Read(newCookieBytes); err != nil {
		//	panic(err)
		//}
		//newCookie := base64.URLEncoding.EncodeToString(newCookieBytes)
		//v.Cookie = newCookie
		//return newCookie, true
		return v.Cookie, true
	} else {
		return "", false
	}
}

func authMiddleware(c *gin.Context) {
	loginCookie, _ := c.Cookie(loginCookieName)
	if loginValidator.check(loginCookie) {
		c.Next()
	} else {
		c.JSON(401, gin.H{
			"code":    8,
			"message": "not logged in",
		})
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
	c.SetSameSite(http.SameSiteStrictMode)
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
		"platform":      runtime.GOOS + "-" + runtime.GOARCH,
	})
}

func initialLoginAuth() {
	cookieBytes := sha256.Sum256(append([]byte(Config.Password), coldSalt...)) //每次运行相同
	loginValidator.Cookie = hex.EncodeToString(cookieBytes[:])
}
