package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	turnstileEndpoint = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
)

type cloudflareResponse struct {
	Success            bool     `json:"success"`
	ErrorCodes         []string `json:"error-codes"`
	ChallengeTimestamp string   `json:"challenge_ts"`
	Hostname           string   `json:"hostname"`
}

type CloudflareService struct {
	turnstileSecret string
	logger          Logger
}

func NewCloudflareService(secret string, logger Logger) *CloudflareService {
	return &CloudflareService{
		turnstileSecret: secret,
		logger:          logger,
	}
}

func (c *CloudflareService) VerifyTurnstile(response string, ip string) (bool, error) {
	body := url.Values{}
	body.Add("secret", c.turnstileSecret)
	body.Add("response", response)
	body.Add("remoteip", ip)

	resp, err := http.PostForm(turnstileEndpoint, body)
	if err != nil {
		c.logger.ERROR(fmt.Sprintf("an error occurred while validating turnstile: %+v", err))
		return false, err
	}

	var respBody cloudflareResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return false, fmt.Errorf("not able to parse cloudflare response")
	}

	if respBody.Success {
		return true, nil
	}

	var cloudflareError string
	for _, v := range respBody.ErrorCodes {
		cloudflareError += fmt.Sprintf("%s,", v)
	}
	err = fmt.Errorf("an invalid status code was returned from turnstile validation: %s: error: %s", resp.Status, cloudflareError)
	c.logger.ERROR(err.Error())
	return false, err
}
