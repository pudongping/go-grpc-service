package bapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	AppKey    = "test"
	AppSecret = "123456"
)

type AccessToken struct {
	Token string `json:"token"`
}

type API struct {
	URL string
}

func NewAPI(url string) *API {
	return &API{URL: url}
}

func (a *API) httpGet(ctx context.Context, path string) ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf("%s/%s", a.URL, path))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

func (a *API) httpPost(ctx context.Context, path, params string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", a.URL, path)
	payload := strings.NewReader(params)
	resp, err := http.Post(url, "application/x-www-form-urlencoded", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

func (a *API) getAccessToken(ctx context.Context) (string, error) {
	// 返回结果类似
	// {"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhcHBfa2V5IjoiMDk4ZjZiY2Q0NjIxZDM3M2NhZGU0ZTgzMjYyN2I0ZjYiLCJhcHBfc2VjcmV0IjoiZTEwYWRjMzk0OWJhNTlhYmJlNTZlMDU3ZjIwZjg4M2UiLCJleHAiOjE2NjU1NzEzODUsImlzcyI6Imdpbi1ibG9nLXNlcnZpY2UifQ.YLT8trySHhiu3S43VcGAQU4fDkrLXsYf9AstSQTKqHA"}
	body, err := a.httpPost(ctx, "auth", fmt.Sprintf("app_key=%s&app_secret=%s", AppKey, AppSecret))
	if err != nil {
		return "", err
	}

	var accessToken AccessToken
	_ = json.Unmarshal(body, &accessToken)
	return accessToken.Token, nil
}

func (a *API) GetTagList(ctx context.Context, name string) ([]byte, error) {
	token, err := a.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	body, err := a.httpGet(ctx, fmt.Sprintf("%s?token=%s&name=%s", "api/v1/tags", token, name))
	if err != nil {
		return nil, err
	}

	return body, nil
}
