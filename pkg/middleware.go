package pkg

import (
	"context"
	"fmt"
	"net/http"

	"github.com/vishu42/terraformer/pkg/logger"
	"github.com/vishu42/terraformer/pkg/oauth"
)

type EnsureAuth struct {
	logHandler http.Handler
	Logger     logger.Logger
	Config     *Config
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

	// parse the token
	_, c, err := oauth.ParseToken(token)
	if err != nil {
		fmt.Println("error parsing token - ", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	fmt.Println("middleware", c)

	// get the user info
	userInfo, err := GetUserInfo(token, ea.Config.ClientSecret)
	if err != nil {
		fmt.Println("error getting user info - ", err)
	}

	fmt.Println("WELCOME", userInfo.Email)
	w.Write([]byte("WELCOME " + userInfo.Email + "\n"))

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// // print the claims
	// fmt.Println(t.Claims)

	// set the claims in the context
	r = r.WithContext(oauth.NewContext(context.Background(), c))

	// set the logger in the context
	r = r.WithContext(logger.NewContext(r.Context(), ea.Logger))

	ea.logHandler.ServeHTTP(w, r)
}

func NewEnsureAuth(config *Config, h http.Handler) *EnsureAuth {
	l, err := logger.New(config.Debug)
	if err != nil {
		panic(err)
	}
	return &EnsureAuth{h, l, config}
}
