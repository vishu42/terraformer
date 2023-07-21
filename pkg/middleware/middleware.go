package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt"
)

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

type EnsureAuth struct {
	logHandler http.Handler
}

func (ea *EnsureAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("hello - I am middleware :)")

	// get the authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("missing authorization header"))
		return
	}

	// get the token
	token := authHeader[len("Bearer "):]

	// get the public key set
	publicKeySet := GetMsPublicKey()

	// verify the token
	// Parse the token without verifying the signature
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
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
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// print the claims
	fmt.Println(t.Claims)

	ea.logHandler.ServeHTTP(w, r)
}

func NewEnsureAuth(handlerToWrap http.Handler) *EnsureAuth {
	return &EnsureAuth{handlerToWrap}
}
