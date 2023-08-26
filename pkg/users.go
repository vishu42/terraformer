package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
)

// this file exposes functions to get user related information
// in exchange of access token

const (
	UserInfoURL = "https://graph.microsoft.com/oidc/userinfo"
	MSTokenURL  = "https://login.microsoftonline.com/4d16a70b-76a1-4ad7-944a-13513528982b/oauth2/v2.0/token"
)

type UserInfo struct {
	Sub        string `json:"sub"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Email      string `json:"email"`
}

type assertionToken struct {
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	ExtExpiresIn int    `json:"ext_expires_in"`
	IdToken      string `json:"id_token"`
}

func getAssertionToken(accessToken, oauthClientSecret string) (token *assertionToken, err error) {
	if accessToken == "" {
		fmt.Println("getassertiontoken - accessToken is empty")
		return
	}

	if oauthClientSecret == "" {
		fmt.Println("getassertiontoken - oauthClientSecret is empty")
		return
	}

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	_ = writer.WriteField("client_id", "292def6c-3c4c-42ed-9ef2-8a197c7abc33")
	_ = writer.WriteField("client_secret", oauthClientSecret)
	_ = writer.WriteField("assertion", accessToken)
	_ = writer.WriteField("scope", "openid profile email User.Read")
	_ = writer.WriteField("requested_token_use", "on_behalf_of")

	err = writer.Close()
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, MSTokenURL, payload)
	if err != nil {
		return
	}

	client := &http.Client{}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// // print req
	// dump, err := httputil.DumpRequest(req, true)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println(string(dump))

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		fmt.Println("getassertiontoken - Status Not OK")
		fmt.Println(res.StatusCode)
	}

	err = json.NewDecoder(res.Body).Decode(&token)
	if err != nil {
		return
	}
	return
}

func GetUserInfo(accessToken, oauthClientSecret string) (userInfo *UserInfo, err error) {
	token, err := getAssertionToken(accessToken, oauthClientSecret)
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, UserInfoURL, nil)
	if err != nil {
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	// print req
	// dump, err := httputil.DumpRequest(req, true)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println(string(dump))
	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		fmt.Println("GetUserInfo - Status Not OK")
		fmt.Println(res.StatusCode)
	}

	err = json.NewDecoder(res.Body).Decode(&userInfo)
	if err != nil {
		return
	}
	return
}
