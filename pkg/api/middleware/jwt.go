package middleware

import (
	"github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
	"time"
	"zeus/pkg/api/domain/account"
	"zeus/pkg/api/dto"
	"zeus/pkg/api/log"
	"zeus/pkg/api/model"
	"zeus/pkg/api/service"
)

var accountService = service.UserService{}
var loginType int

//todo : 用单独的claims model去掉user model
func JwtAuth(loginType int) *jwt.GinJWTMiddleware {
	jwtMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:            "Jwt",
		SigningAlgorithm: "RS256",
		PubKeyFile:       viper.GetString("jwt.key.public"),
		PrivKeyFile:      viper.GetString("jwt.key.private"),
		Timeout:          time.Hour * 24,
		MaxRefresh:       time.Hour * 24 * 90,
		IdentityKey:      "id",
		LoginResponse:    LoginResponse,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(model.UserClaims); ok {
				return jwt.MapClaims{
					"id":   v.Id,
					"name": v.Name,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return model.UserClaims{
				Name: claims["name"].(string),
				Id:   int(claims["id"].(float64)),
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			if loginType == account.LoginStandard.Type {
				return Authenticator(c)
			}
			return AuthenticatorOAuth(c)
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			if _, ok := data.(model.UserClaims); ok {
				return true
			}
			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code": code,
				"msg":  message,
			})
		},
		TokenLookup:   "header: Authorization, query: token, cookie: jwt",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	})
	if err != nil {
		log.Error(err.Error())
	}
	return jwtMiddleware
}

func LoginResponse(c *gin.Context, code int, token string, expire time.Time) {
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"data": map[string]interface{}{
			"token":  token,
			"expire": expire,
		},
	})
}
func Authenticator(c *gin.Context) (interface{}, error) {
	var loginDto dto.LoginDto
	if err := dto.Bind(c, &loginDto); err != nil {
		return "", err
	}
	ok, u := accountService.VerifyAndReturnUserInfo(loginDto)
	if ok {
		return model.UserClaims{
			Id:   u.Id,
			Name: u.Username,
		}, nil
	}
	return nil, jwt.ErrFailedAuthentication
}

func AuthenticatorOAuth(c *gin.Context) (interface{}, error) {
	oauthDto := &dto.LoginOAuthDto{}
	if err := dto.Bind(c, &oauthDto); err != nil {
		return "", err
	}
	//TODO 支持微信、钉钉、QQ等登陆
	userOauth, err := accountService.VerifyDTAndReturnUserInfo(oauthDto.Code)
	if err != nil || userOauth == nil {
		return "", err
	}
	return model.UserClaims{
		Id:   userOauth.Id,
		Name: userOauth.Name,
	}, nil
}
