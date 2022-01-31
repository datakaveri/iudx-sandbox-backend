package authutility

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
)

type User struct {
	UserName string
}

func extractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func verifyToken(r *http.Request) (*jwt.Token, error) {
	SecretKey := "-----BEGIN CERTIFICATE-----\n" +
		"MIIClzCCAX8CBgF6vZa5XjANBgkqhkiG9w0BAQsFADAPMQ0wCwYDVQQDDARkZW1v" +
		"MB4XDTIxMDcxOTA3MDUwOVoXDTMxMDcxOTA3MDY0OVowDzENMAsGA1UEAwwEZGVt" +
		"bzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKi1f2WK81JBxFlxuwps" +
		"jhj6ws32LPbYbDrsGO2qcXne+bxHro02sow4dMW0I+Lff5o3aqc3CRRxAniWfr+Z" +
		"ttaRpLHC5n5wJxWznq0lEFOdtr7a0+mZXflLPB9HLCOzeA1Vfiat44BM1OiJ5kUJ" +
		"AQmiA+9Ww9rKkVCoUMpBkCdK0a9uBi8pQnQVY76GlSez0k5wO8zsRWZ4b+P90DtV" +
		"+tscluoAFopSy48vt63f76i0/Mi9if7wHpP12wGFF85Z3U75skafwVf3bcRLGsLI" +
		"Z05GERojp1oV+Hr0ob4ayG2UrrT1Ama2l2/QfS2aNOwt/F5S2l2RFCFfUZqv3muz" +
		"f38CAwEAATANBgkqhkiG9w0BAQsFAAOCAQEAVcs+sFMmhUlyg7d9C2gi2vOBXiBs" +
		"MVd1Xm0WgxZY1ISwAEj4tr9SALsHw8N/gWzL7uYOWJeNpvCb63/HnC6AXhTRJWkq" +
		"jb+SqPPXhS2dKLiMAicFaPIQewfqqIR6k9kiwbWTiY8Xy00xSlznNx69OUQis3Zc" +
		"k/uQJaEBgvJQ5PfLGWn0yrjF2REW/bdPDnoW5qBX7jHmg2fFxdqbean9w+G+xMQX" +
		"L4FWl+jddcmMCyhg3IblYfW0rsCWpRFShXl535awvmtXAiDB7weD0Nd/FbO5Us/u" +
		"lyBP1aC/MuuncHcG7oDPA4Ao++3de8xXk0jdDxczvRYJIjLa3q5EwKfo5w==" +
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
