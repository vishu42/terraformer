package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt"
)

type MyCustomClaims struct {
	Roles []string `json:"roles"`
	jwt.StandardClaims
}

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type key int

// oauthkey is the context key for the user IP address.  Its value of zero is
// arbitrary.  If this package defined other context keys, they would have
// different integer values.
const oauthKey key = 0

type PublicKeySet struct {
	Keys []PublicKey `json:"keys"`
}
type PublicKey struct {
	KTY    string   `json:"kty"`
	KID    string   `json:"kid"`
	Use    string   `json:"use"`
	N      string   `json:"n"`
	E      string   `json:"e"`
	X5C    []string `json:"x5c"`
	Issuer string   `json:"issuer"`
}

func GetMsPublicKey() PublicKeySet {
	microsoftKeysURL := "https://login.microsoftonline.com/common/discovery/v2.0/keys"

	client := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, microsoftKeysURL, nil)
	if err != nil {
		log.Fatalf("unable to generate http req - %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error executing http req - %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatal("Status Not OK")
	}

	defer resp.Body.Close()

	if err != nil {
		log.Fatalf("error closing resp body - %v", err)
	}

	var publicKeys PublicKeySet

	err = json.NewDecoder(resp.Body).Decode(&publicKeys)
	if err != nil {
		log.Fatalf("error decoding resp body - %v", err)
	}

	return publicKeys
}

func ParseToken(token string) (*jwt.Token, *MyCustomClaims, error) {
	// get the public key set
	publicKeySet := GetMsPublicKey()

	// Parse the token without verifying the signature
	t, err := jwt.ParseWithClaims(token, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// find the public key
		var key string
		for _, v := range publicKeySet.Keys {
			if v.KID == token.Header["kid"] {
				key = v.X5C[0]
				break
			}
		}

		// embed the public key in the PEM format
		pem := "-----BEGIN CERTIFICATE-----\n" + key + "\n-----END CERTIFICATE-----"

		// parse the PEM encoded public key
		result, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pem))
		if err != nil {
			return nil, err
		}

		return result, nil
	})
	if err != nil {
		return nil, nil, err
	}

	claims, ok := t.Claims.(*MyCustomClaims)
	if !ok || !t.Valid {
		return nil, nil, fmt.Errorf("invalid token")
	}

	// print claims
	fmt.Println(claims)

	return t, claims, nil
}

func NewContext(ctx context.Context, claims *MyCustomClaims) context.Context {
	return context.WithValue(ctx, oauthKey, claims)
}

func FromContext(ctx context.Context) (*MyCustomClaims, bool) {
	claims, ok := ctx.Value(oauthKey).(*MyCustomClaims)
	fmt.Println(claims)
	return claims, ok
}
