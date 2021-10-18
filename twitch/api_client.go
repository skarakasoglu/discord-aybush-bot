package twitch

import (
	"encoding/json"
	"fmt"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/payloads"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/payloads/v1"
	v2 "github.com/skarakasoglu/discord-aybush-bot/twitch/payloads/v2"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var (
	BASE_AUTH_URL  = "https://id.twitch.tv"
	AUTH_VERSION   = "oauth2"
	TOKEN_ENDPOINT = "token"

	BASE_API_URL = "https://api.twitch.tv"
	API_VERSION = "helix"
	USERS_ENDPOINT = "users"
	FOLLOWS_ENDPOINT = USERS_ENDPOINT + "/follows"
	GAMES_ENDPOINT = "games"
	STREAMS_ENDPOINT = "streams"
	CHANNELS_ENDPOINT = "channels"
	EVENTSUB_ENDPOINT = "eventsub"
	SUBSCRIPTIONS_ENDPOINT = "subscriptions"

)

type ApiClient struct {
	appAccessToken string
	userAccessToken string

	userRefreshToken string

	clientId string
	clientSecret string
	authorizationCode string
	redirectUri string
}

func NewApiClient(clientId string, clientSecret string, userRefreshToken string) *ApiClient{
	api := &ApiClient{
		clientId:         clientId,
		clientSecret:     clientSecret,
		userRefreshToken: userRefreshToken,
	}
	api.generateAppAccessToken()
	api.refreshUserAccessToken()

	return api
}

func (api *ApiClient) refreshUserAccessToken() payloads.AccessToken {
	var accessToken payloads.AccessToken

	reqUrl := fmt.Sprintf("%v/%v/%v", BASE_AUTH_URL, AUTH_VERSION, TOKEN_ENDPOINT)

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", api.userRefreshToken)
	data.Set("client_id", api.clientId)
	data.Set("client_secret", api.clientSecret)

	req, err := http.NewRequest(http.MethodPost, reqUrl, strings.NewReader(data.Encode()))
	if err != nil {
		log.Printf("[TwitchApiClient] Error on creating new request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on making request: %v", err)
		return accessToken
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on reading response body: %v", err)
	}

	err = json.Unmarshal(buffer, &accessToken)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on unmarshalling json: %v", err)
		return accessToken
	}

	if resp.StatusCode == http.StatusUnauthorized {
		log.Printf("[TwitchApiClient] Error on generating user access token: %v", string(buffer))
		return accessToken
	}

	log.Printf("[TwitchApiClient] Twitch user access token generated successfully. Response: %v", string(buffer))
	api.userAccessToken = accessToken.AccessToken
	api.userRefreshToken = accessToken.RefreshToken

	return accessToken
}

//DEPRECATED
func (api *ApiClient) generateUserAccessToken() payloads.AccessToken {
	var token payloads.AccessToken

	reqUrl := fmt.Sprintf("%v/%v/%v?client_id=%v&client_secret=%v&code=%v&grant_type=authorization_code&redirect_uri=%v",
		BASE_AUTH_URL, AUTH_VERSION, TOKEN_ENDPOINT,
		api.clientId, api.clientSecret, api.authorizationCode, api.redirectUri)
	req, err := http.NewRequest(http.MethodPost, reqUrl, nil)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on creating new request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on making request: %v", err)
	return token
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on reading response body: %v", err)
	}

	var accessToken payloads.AccessToken
	err = json.Unmarshal(buffer, &accessToken)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on unmarshalling json: %v", err)
		return token
	}

	if resp.StatusCode == http.StatusUnauthorized {
		log.Printf("[TwitchApiClient] Error on generating user access token: %v", string(buffer))
		return token
	}

	log.Printf("[TwitchApiClient] Twitch user access token generated successfully. Response: %v", string(buffer))
	api.userAccessToken = accessToken.AccessToken
	api.userRefreshToken = accessToken.RefreshToken

	return accessToken
}

func (api *ApiClient) generateAppAccessToken() {
	reqUrl := fmt.Sprintf("%v/%v/%v?client_id=%v&client_secret=%v&grant_type=client_credentials",
		BASE_AUTH_URL, AUTH_VERSION, TOKEN_ENDPOINT,
		api.clientId, api.clientSecret)
	req, err := http.NewRequest(http.MethodPost, reqUrl, nil)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on creating new request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on making request: %v", err)
		return
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on reading response body: %v", err)
	}

	var accessToken payloads.AccessToken
	err = json.Unmarshal(buffer, &accessToken)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on unmarshalling json: %v", err)
		return
	}

	if resp.StatusCode == http.StatusUnauthorized {
		log.Printf("[TwitchApiClient] Error on generating app access token: %v", string(buffer))
		return
	}

	log.Printf("[TwitchApiClient] Twitch Api access token generated successfully. Response: %v", string(buffer))
	api.appAccessToken = accessToken.AccessToken
}

func (api *ApiClient) getUserFollowage(fromId string, toId string) v1.UserFollows {
	followageReqUrl := fmt.Sprintf("%v/%v/%v?from_id=%v&to_id=%v",
		BASE_API_URL, API_VERSION, FOLLOWS_ENDPOINT,
		fromId, toId)
	resp, err := api.makeHttpGetRequest(followageReqUrl)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on making request: %v", err)
		return v1.UserFollows{}
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on reading response body: %v", err)
		return v1.UserFollows{}
	}

	var followPayload v1.UserFollowsPayload
	err = json.Unmarshal(buffer, &followPayload)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on unmarshalling json: %v", err)
		return v1.UserFollows{}
	}

	if len(followPayload.Data) < 1 {
		return v1.UserFollows{}
	}

	return followPayload.Data[0]
}

func (api *ApiClient) getUserInfoByUserId(userId string) payloads.User {
	reqUrl := fmt.Sprintf("%v/%v/%v?id=%v", BASE_API_URL, API_VERSION, USERS_ENDPOINT, userId)
	return api.getUserInfo(reqUrl)
}

func (api *ApiClient) getUserInfoByUsername(username string) payloads.User {
	reqUrl := fmt.Sprintf("%v/%v/%v?login=%v", BASE_API_URL, API_VERSION, USERS_ENDPOINT, username)
	return api.getUserInfo(reqUrl)
}

func (api *ApiClient) getUserInfo(reqUrl string) payloads.User {
	resp, err := api.makeHttpGetRequest(reqUrl)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on making request: %v", err)
		return payloads.User{}
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on reading response body: %v", err)
		return payloads.User{}
	}

	var userPayload payloads.UserPayload
	err = json.Unmarshal(buffer, &userPayload)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on unmarshalling json: %v", err)
	}

	if len(userPayload.Data) < 1 {
		log.Printf("No streamer found.")
		return payloads.User{}
	}

	return userPayload.Data[0]
}

func (api *ApiClient) getChannelInfoByBroadcasterId(broadcasterId string) payloads.ChannelInfo {
	channelReqUrl := fmt.Sprintf("%v/%v/%v?broadcaster_id=%v", BASE_API_URL, API_VERSION, GAMES_ENDPOINT, broadcasterId)

	resp, err := api.makeHttpGetRequest(channelReqUrl)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on making request: %v", err)
		return payloads.ChannelInfo{}
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on reading response body: %v", err)
		return payloads.ChannelInfo{}
	}

	var channelInfo payloads.ChannelPayload
	err = json.Unmarshal(buffer, &channelInfo)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on unmarshalling json: %v", err)
	}

	if len(channelInfo.Data) < 1 {
		log.Printf("No streamer found.")
		return payloads.ChannelInfo{}
	}

	return channelInfo.Data[0]
}

func (api *ApiClient) getGameById(gameID string) payloads.Game {
	gameReqUrl := fmt.Sprintf("%v/%v/%v?id=%v", BASE_API_URL, API_VERSION, GAMES_ENDPOINT, gameID)
	resp, err := api.makeHttpGetRequest(gameReqUrl)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on making request: %v", err)
		return payloads.Game{}
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on reading response body: %v", err)
		return payloads.Game{}
	}

	var gamePayload payloads.GamePayload
	err = json.Unmarshal(buffer, &gamePayload)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on unmarshalling json: %v", err)
	}

	if len(gamePayload.Data) < 1 {
		log.Printf("[TwitchApiClient] No game found.")
		return payloads.Game{}
	}

	return gamePayload.Data[0]
}

func (api *ApiClient) getAllSubscriptions() []v2.Subscription {
	var subscriptions []v2.Subscription

	subscriptionsUrl := fmt.Sprintf("%v/%v/%v/%v", BASE_API_URL, API_VERSION, EVENTSUB_ENDPOINT, SUBSCRIPTIONS_ENDPOINT)
	resp, err := api.makeHttpGetRequest(subscriptionsUrl)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on making request: %v", err)
		return subscriptions
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on reading response body: %v", err)
		return subscriptions
	}

	var payload v2.SubscriptionPayload
	err = json.Unmarshal(buffer, &payload)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on unmarshalling json: %v", err)
	}

	if len(payload.Data) < 1 {
		log.Printf("[TwitchApiClient] No subscription found.")
		return subscriptions
	}

	return payload.Data
}

func (api *ApiClient) deleteSubscription(subscription v2.Subscription) {
	subscriptionsUrl := fmt.Sprintf("%v/%v/%v/%v?id=%v", BASE_API_URL, API_VERSION, EVENTSUB_ENDPOINT, SUBSCRIPTIONS_ENDPOINT, subscription.Id)
	_, err := api.makeHttpRequest(subscriptionsUrl, http.MethodDelete)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on making request: %v", err)
		return
	}

	log.Printf("[TwitchApiClient] %v subscription id unsubscribed from %v event successfully.", subscription.Id, subscription.Type)
}

func (api *ApiClient) makeHttpGetRequest(requestURL string) (*http.Response, error) {
	return api.makeHttpRequest(requestURL, http.MethodGet)
}


func (api *ApiClient) makeHttpRequest(requestURL string, method string) (*http.Response, error) {
	var resp *http.Response
	var err error

	tokenExpired := false

	for {
		req, err := http.NewRequest(method, requestURL, nil)

		if err != nil {
			log.Printf("[TwitchApiClient] Error on creating new request: %v", req)
			return nil, err
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", api.appAccessToken))
		req.Header.Set("Client-ID", api.clientId)

		client := &http.Client{}
		resp, err = client.Do(req)
		if err != nil {
			log.Printf("[TwitchApiClient] Error on making request: %v", err)
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