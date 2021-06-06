package streamlabs

import (
	"encoding/json"
	"fmt"
	"github.com/skarakasoglu/discord-aybush-bot/streamlabs/models"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const (
	BASE_API_URL = "https://streamlabs.com/api"
	API_VER = "v1.0"

	DONATIONS_ENDPOINT = "donations"
	ALERTS_ENDPOINT = "alerts"
)

type ApiClient struct{
	accessToken string
}

func NewApiClient(accessToken string) ApiClient{
	return ApiClient{accessToken: accessToken}
}

func (api ApiClient) GetDonations() ([]models.GetDonation, error) {
	var donations models.GetDonationResponse

	requestURL := fmt.Sprintf("%v/%v/%v?access_token=%v", BASE_API_URL, API_VER, DONATIONS_ENDPOINT, api.accessToken)

	resp, err := api.makeRequest(http.MethodGet, requestURL, nil, nil)
	if err != nil {
		log.Printf("[StreamLabsApiClient] Error on making request: %v", err)
		return donations.Data, err
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[StreamLabsApiClient] Error on reading response body: %v", err)
		return donations.Data, err
	}

	err = json.Unmarshal(buffer, &donations)
	if err != nil {
		log.Printf("[StreamLabsApiClient] Error on unmarshalling to JSON: %v", err)
		return donations.Data, err
	}

	return donations.Data, nil
}

func (api ApiClient) CreateDonation(donation models.CreateDonation) (int64, error) {
	var donationResponse models.CreateDonationResponse

	requestURL := fmt.Sprintf("%v/%v/%v", BASE_API_URL, API_VER, DONATIONS_ENDPOINT)

	params := url.Values{}
	params.Set("access_token", api.accessToken)
	params.Set("name", donation.Name)
	params.Set("message", donation.Message)
	params.Set("identifier", donation.Identifier)
	params.Set("currency", donation.Currency)
	//params.Set("created_at", fmt.Sprint(donation.CreatedAt))
	params.Set("amount", fmt.Sprint(donation.Amount))
	params.Set("skip_alert", donation.SkipAlert)

	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	resp, err := api.makeRequest(http.MethodPost, requestURL, headers, strings.NewReader(params.Encode()))
	if err != nil {
		log.Printf("[StreamLabsApiClient] Error on making request: %v", err)
		return donationResponse.DonationId, err
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[StreamLabsApiClient] Error on reading response body: %v", err)
		return donationResponse.DonationId, err
	}

	err = json.Unmarshal(buffer, &donationResponse)
	if err != nil {
		log.Printf("[StreamLabsApiClient] Error on unmarshalling to JSON: %v", err)
		return donationResponse.DonationId, err
	}

	if resp.StatusCode != http.StatusOK {
		var errResp models.Error
		err = json.Unmarshal(buffer, &errResp)
		if err != nil {
			log.Printf("[StreamLabsApiClient] Error on unmarshalling to JSON: %v", err)
			return donationResponse.DonationId, err
		}

		log.Printf("[StreamLabsApiClient] Error on response: %v, message: %v", errResp.Error, errResp.Message)
	}

	return donationResponse.DonationId, nil
}

func (api ApiClient) CreateAlert(alert models.CreateAlert) (bool, error) {
	var response models.CreateAlertResponse

	requestURL := fmt.Sprintf("%v/%v/%v", BASE_API_URL, API_VER, ALERTS_ENDPOINT)

	params := url.Values{}
	params.Set("access_token", api.accessToken)
	params.Set("type", alert.Type.String())
	params.Set("image_href", alert.ImageHref)
	params.Set("sound_href", alert.SoundHref)
	params.Set("message", alert.Message)
	params.Set("user_message", alert.UserMessage)
	params.Set("duration", fmt.Sprint(alert.Duration))
	params.Set("special_text_color", alert.SpecialTextColor)

	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	resp, err := api.makeRequest(http.MethodPost, requestURL, headers, strings.NewReader(params.Encode()))
	if err != nil {
		log.Printf("[StreamLabsApiClient] Error on making request: %v", err)
		return false, err
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[StreamLabsApiClient] Error on reading response body: %v", err)
		return false, err
	}

	err = json.Unmarshal(buffer, &response)
	if err != nil {
		log.Printf("[StreamLabsApiClient] Error on unmarshalling to JSON: %v", err)
		return false, err
	}

	if resp.StatusCode != http.StatusOK {
		var errResp models.Error
		err = json.Unmarshal(buffer, &errResp)
		if err != nil {
			log.Printf("[StreamLabsApiClient] Error on unmarshalling to JSON: %v", err)
			return false, err
		}

		log.Printf("[StreamLabsApiClient] Error on response: %v, message: %v", errResp.Error, errResp.Message)
	}

	return response.Success, nil
}

func (api ApiClient) makeRequest(method string, requestURL string, headers map[string]string, body io.Reader) (*http.Response, error) {
	var resp *http.Response

	req, err := http.NewRequest(method, requestURL, body)
	if err != nil {
		log.Printf("[StreamLabsApiClient] Error on creating new request: %v", req)
		return resp, err
	}

	if headers != nil {
		for k, val := range headers {
			req.Header.Set(k, val)
		}
	}

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		log.Printf("[StreamLabsApiClient] Error on making request: %v", err)
		return resp, err
	}

	return resp, nil
}