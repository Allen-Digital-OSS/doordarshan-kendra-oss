package common

import (
	"errors"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/constant"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/gommon/log"
)

const (
	ClaimUserID      = "uid"
	ClaimPersonaType = "pt"
	ClaimTenantID    = "tid"
	ClaimExternalID  = "e_id"
)

var (
	ErrUnauthorizedEmptyToken       = errors.New("token is empty")
	ErrUnauthorizedInvalidToken     = errors.New("token is invalid")
	ErrUnauthorizedInvalidSignature = errors.New("signature is invalid")
)

type UserInfoFromJWTClaims struct {
	UserID      string `json:"uid"`
	PersonaType string `json:"pt"`
	TenantID    string `json:"tid"`
	ExternalID  string `json:"e_id"`
}

// ExtractClaimsFromToken parses the JWT token and extracts the expected claims.
func ExtractClaimsFromToken(tokenStr, jwtSecret string) (*UserInfoFromJWTClaims, error) {
	claimsMap := make(map[string]string)
	if tokenStr == constant.EmptyStr {
		return nil, ErrUnauthorizedEmptyToken
	}

	// Parse the token
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnauthorizedInvalidSignature
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		log.Errorf("[JWT] Error while parsing token err: %s", err)
		return nil, ErrUnauthorizedInvalidToken
	}

	// Validate the token and extract the claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		claimsRequired := getClaimsToBeExtracted()
		for _, claimKey := range claimsRequired {
			claimsMap[claimKey], ok = claims[claimKey].(string)
			if !ok {
				return nil, ErrUnauthorizedInvalidToken
			}
		}

		if _, ok = claimsMap[ClaimUserID]; !ok {
			log.Errorf("[JWT] UserID not found err: %s", err)
			return nil, ErrUnauthorizedInvalidToken
		}
		if _, ok = claimsMap[ClaimPersonaType]; !ok {
			log.Errorf("[JWT] PersonaType not found, err: %s", err)
			return nil, ErrUnauthorizedInvalidToken
		}
		if _, ok = claimsMap[ClaimTenantID]; !ok {
			log.Errorf("[JWT] TenantID not found, err: %s", err)
			return nil, ErrUnauthorizedInvalidToken
		}
		if _, ok = claimsMap[ClaimExternalID]; !ok {
			log.Errorf("[JWT] External ID not found, err: %s", err)
			return nil, ErrUnauthorizedInvalidToken
		}

		return &UserInfoFromJWTClaims{
			UserID:      claimsMap[ClaimUserID],
			PersonaType: claimsMap[ClaimPersonaType],
			TenantID:    claimsMap[ClaimTenantID],
			ExternalID:  claimsMap[ClaimExternalID],
		}, nil
	}

	return nil, ErrUnauthorizedInvalidToken
}

func getClaimsToBeExtracted() []string {
	return []string{
		ClaimUserID,
		ClaimPersonaType,
		ClaimTenantID,
		ClaimExternalID,
	}
}
