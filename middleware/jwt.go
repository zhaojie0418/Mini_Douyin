package middleware

import (
	"github.com/ACking-you/byte_douyin_project/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

//签名所需混淆密钥 不要太简单 容易被破解
var jwtKey = []byte("acking-you.xyz")

type Claims struct {
	UserId int64
	jwt.StandardClaims
}

// ReleaseToken 颁发token 只有在用户登录过程中才对其发放token
func ReleaseToken(user models.UserLogin) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour)//设定token的过期时间
	claims := &Claims{
		UserId: user.UserInfoId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "douyin_pro_131", //token的签发者
			Subject:   "L_B__",
		}}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)//使用给定的加密算法生成对应token
	tokenString, err := token.SignedString(jwtKey)//SignedString 方法根据传入的空接口类型参数 key，返回完整的签名令牌。
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// ParseToken 解析token
func ParseToken(tokenString string) (*Claims, bool) {
	token, _ := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if token != nil {
		if key, ok := token.Claims.(*Claims); ok {
			if token.Valid {
				return key, true
			} else {
				return key, false
			}
		}
	}
	return nil, false
}

// JWTMiddleWare 鉴权中间件，鉴权并设置user_id
func JWTMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.Query("token")
		if tokenStr == "" {
			tokenStr = c.PostForm("token")
		}
		//用户不存在
		if tokenStr == "" {
			c.JSON(http.StatusOK, models.CommonResponse{StatusCode: 401, StatusMsg: "用户不存在"})
			c.Abort() //阻止执行
			return
		}
		//验证token
		tokenStruck, ok := ParseToken(tokenStr)
		if !ok {
			c.JSON(http.StatusOK, models.CommonResponse{
				StatusCode: 403,
				StatusMsg:  "token不正确",
			})
			c.Abort() //阻止执行
			return
		}
		//token超时
		if time.Now().Unix() > tokenStruck.ExpiresAt {
			c.JSON(http.StatusOK, models.CommonResponse{
				StatusCode: 402,
				StatusMsg:  "token过期",
			})
			c.Abort() //阻止执行
			return
		}
		c.Set("user_id", tokenStruck.UserId)
		c.Next()
	}
}
