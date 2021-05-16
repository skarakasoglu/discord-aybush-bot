package twitch

import (
	"encoding/json"
	"fmt"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/payloads"
	"io/ioutil"
	"log"
	"net/http"
)

type ApiClient struct {
	appAccessToken string
	userAccessToken string

	clientId string
	clientSecret string
	authorizationCode string
	redirectUri string
}

func NewApiClient(clientId string, clientSecret string, authorizationCode string, redirectUri string) *ApiClient{
	api := &ApiClient{
		clientId:         clientId,
		clientSecret:     clientSecret,
		authorizationCode: authorizationCode,
		redirectUri: redirectUri,
	}
	api.generateAppAccessToken()
	api.generateUserAccessToken()

	return api
}

func (api *ApiClient) generateUserAccessToken() {
	reqUrl := fmt.Sprintf("https://id.twitch.tv/oauth2/token?client_id=%v&client_secret=%v&code=%v&grant_type=authorization_code&redirect_uri=%v", api.clientId, api.clientSecret, api.authorizationCode, api.redirectUri)
	req, err := http.NewRequest(http.MethodPost, reqUrl, nil)
	if err != nil {
		log.Printf("Error on creating new request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error on making request: %v", err)
	return
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error on reading response body: %v", err)
	}

	var accessToken payloads.AccessToken
	err = json.Unmarshal(buffer, &accessToken)
	if err != nil {
		log.Printf("Error on unmarshalling json: %v", err)
	return
	}

	log.Printf("Twitch api access token generated successfully. Token: %v, Type: %v", accessToken.AccessToken, accessToken.TokenType)
	api.userAccessToken = accessToken.AccessToken
}

func (api *ApiClient) generateAppAccessToken() {
	reqUrl := fmt.Sprintf("https://id.twitch.tv/oauth2/token?client_id=%v&client_secret=%v&grant_type=client_credentials", api.clientId, api.clientSecret)
	req, err := http.NewRequest(http.MethodPost, reqUrl, nil)
	if err != nil {
		log.Printf("Error on creating new request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error on making request: %v", err)
		return
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error on reading response body: %v", err)
	}

	var accessToken payloads.AccessToken
	err = json.Unmarshal(buffer, &accessToken)
	if err != nil {
		log.Printf("Error on unmarshalling json: %v", err)
		return
	}

	log.Printf("Twitch api access token generated successfully. Token: %v, Type: %v", accessToken.AccessToken, accessToken.TokenType)
	api.appAccessToken = accessToken.AccessToken
}

func (api *ApiClient) getUserFollowage(fromId int, toId int) payloads.UserFollows {
	followageReqUrl := fmt.Sprintf("", fromId, toId)
	resp, err := api.makeHttpGetRequest(followageReqUrl)
	if err != nil {
		log.Printf("Error on making request: %v", err)
		return payloads.UserFollows{}
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error on reading response body: %v", err)
		return payloads.UserFollows{}
	}

	var followPayload payloads.UserFollowsPayload
	err = json.Unmarshal(buffer, &followPayload)
	if err != nil {
		log.Printf("Error on unmarshalling json: %v", err)
		return payloads.UserFollows{}
	}

	if len(followPayload.Data) < 1 {
		return payloads.UserFollows{}
	}

	return followPayload.Data[0]
}

func (api *ApiClient) getUserInfoByUsername(username string) payloads.User {
	gameReqUrl := fmt.Sprintf("https://api.twitch.tv/helix/users?login=%v", username)
	resp, err := api.makeHttpGetRequest(gameReqUrl)
	if err != nil {
		log.Printf("Error on making request: %v", err)
		return payloads.User{}
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error on reading response body: %v", err)
		return payloads.User{}
	}

	var userPayload payloads.UserPayload
	err = json.Unmarshal(buffer, &userPayload)
	if err != nil {
		log.Printf("Error on unmarshalling json: %v", err)
	}

	if len(userPayload.Data) < 1 {
		log.Printf("No streamer found.")
		return payloads.User{}
	}

	return userPayload.Data[0]
}

func (api *ApiClient) getGameById(gameID string) payloads.Game {
	gameReqUrl := fmt.Sprintf("https://api.twitch.tv/helix/games?id=%v", gameID)
	resp, err := api.makeHttpGetRequest(gameReqUrl)
	if err != nil {
		log.Printf("Error on making request: %v", err)
		return payloads.Game{}
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error on reading response body: %v", err)
		return payloads.Game{}
	}

	var gamePayload payloads.GamePayload
	err = json.Unmarshal(buffer, &gamePayload)
	if err != nil {
		log.Printf("Error on unmarshalling json: %v", err)
	}

	if len(gamePayload.Data) < 1 {
		log.Printf("No game found.")
		return payloads.Game{}
	}

	return gamePayload.Data[0]
}

func (api *ApiClient) makeHttpGetRequest(requestURL string) (*http.Response, error) {
	var resp *http.Response
	var err error

	tokenExpired := false

	for {
		req, err := http.NewRequest(http.MethodGet, requestURL, nil)

		if err != nil {
			log.Printf("Error on creating new request: %v", req)
			return nil, err
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", api.appAccessToken))
		req.Header.Set("Client-ID", api.clientId)

		client := &http.Client{}
		resp, err = client.Do(req)
		if err != nil {
			log.Printf("Error on making request: %v", err)
			return nil, err
		}

		if resp.StatusCode == http.StatusUnauthorized {
			tokenExpired = true
			api.generateAppAccessToken()
		} else {
			tokenExpired = false
		}

		if !tokenExpired {
			break
		}
	}

	return resp, err
}