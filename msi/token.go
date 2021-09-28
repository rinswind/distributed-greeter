package msi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	msiEndpoint   = "http://169.254.169.254/metadata/identity/oauth2/token"
	msiApiVersion = "2018-02-01"
)

/*
 * AccessTokenResponse
 *
 * An https://www.oauth.com/oauth2-servers/access-tokens/access-token-response/
 */
type AccessTokenResponse struct {
	// Required
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   string `json:"expires_in"`

	// Optional
	RefreshToken string `json:"refresh_token"`

	// Azure extensions?
	ExpiresOn string `json:"expires_on"`
	NotBefore string `json:"not_before"`
	Resource  string `json:"resource"`
}

type AccessTokenClient struct {
	httpClient *http.Client
	requestUrl string
}

func (atr *AccessTokenResponse) String() string {
	return fmt.Sprintf("{resource: %v, type: %v, expiresOn: %v}", atr.Resource, atr.TokenType, atr.ExpiresOn)
}

func NewAccessTokenClient(resource string) *AccessTokenClient {
	return &AccessTokenClient{
		&http.Client{},
		fmt.Sprintf("%v?api-version=%v&resource=%v", msiEndpoint, msiApiVersion, url.QueryEscape(resource)),
	}
}

func (msi *AccessTokenClient) RequestToken() (*AccessTokenResponse, error) {
	// Create an HTTP request for an Access Token to access the specified resource type
	req, err := http.NewRequest("GET", msi.requestUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Metadata", "true")

	// Call managed services for Azure resources token endpoint
	resp, err := msi.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to obtain access token: request: %v, response: %v (%v)", msi.requestUrl, resp.StatusCode, resp.Status)
	}

	// Pull out response body
	respRaw, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	// Unmarshall response body into struct
	token := AccessTokenResponse{}
	err = json.Unmarshal(respRaw, &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}
