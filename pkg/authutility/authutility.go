package authutility

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt"
)

type User struct {
	UserName string
}

func extractToken(r *http.Request) string {
	bearToken := r.Header.Get("token")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func verifyToken(r *http.Request) (*jwt.Token, error) {
	SecretKey := "-----BEGIN CERTIFICATE-----\n" +
		os.Getenv("KEYCLOAK_PUBLIC_KEY") +
		"\n-----END CERTIFICATE-----"

	key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(SecretKey))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	tokenString := extractToken(r)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return token, nil
}

func TokenValid(r *http.Request) error {
	token, err := verifyToken(r)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok || !token.Valid {
		return err
	}
	return nil
}

func ExtractTokenMetadata(r *http.Request) (*User, error) {
	token, err := verifyToken(r)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		// accessUuid, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, err
		}
		userName := fmt.Sprintf("%s", claims["preferred_username"])
		if err != nil {
			return nil, err
		}
		return &User{
			UserName: userName,
		}, nil
	}
	return nil, err
}
