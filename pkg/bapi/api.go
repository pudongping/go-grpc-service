package bapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"golang.org/x/net/context/ctxhttp"
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

func (a *API) _httpGet(ctx context.Context, path string) ([]byte, error) {
	resp, err := ctxhttp.Get(ctx, http.DefaultClient, fmt.Sprintf("%s/%s", a.URL, path))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

func (a *API) httpGet(ctx context.Context, path string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", a.URL, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	span, _ := opentracing.StartSpanFromContext(
		ctx,
		"HTTP GET: "+a.URL,
		opentracing.Tag{Key: string(ext.Component), Value: "HTTP"},
	)
	span.SetTag("url", url)
	_ = opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)

	req = req.WithContext(context.Background())
	client := http.Client{
		Timeout: time.Second * 60,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	defer span.Finish()

	// 读取消息主体，在实际封装中可以将其抽离
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
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

// 如果需要演示 http 请求的话呢，还需要将 [gin-blog-service](https://github.com/pudongping/gin-blog-service) 项目跑起来才行
func (a *API) _GetTagList(ctx context.Context, name string) ([]byte, error) {
	token, err := a.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	body, err := a.httpGet(ctx, fmt.Sprintf("%s?token=%s&name=%s", "api/v1/tags", token, name))
	if err != nil {
		return nil, err
	}
	fmt.Printf("\n")
	fmt.Println(string(body))

	return body, nil
}

// 为了单纯跑本项目，不依赖其他的项目，故此操作，写一个假数据
// 本项目的重点在于学习 grpc
func (a *API) GetTagList(ctx context.Context, name string) ([]byte, error) {
	resp := `{"list":[{"id":1,"created_on":0,"created_by":"alex","modified_on":1641743745,"modified_by":"alex","deleted_on":0,"is_del":0,"name":"Go","state":1},{"id":2,"created_on":1641719964,"created_by":"alex","modified_on":1641719964,"modified_by":"","deleted_on":0,"is_del":0,"name":"PHP","state":1},{"id":3,"created_on":1641720152,"created_by":"alex","modified_on":1641720152,"modified_by":"","deleted_on":0,"is_del":0,"name":"Rust","state":1}],"pager":{"page":1,"page_size":10,"total_rows":3}}`
	if name == "" {
		return []byte(resp), nil
	}

	res := `{"list":[{"id":1,"created_on":0,"created_by":"alex","modified_on":1641743745,"modified_by":"alex","deleted_on":0,"is_del":0,"name":"` + name + `","state":1}],"pager":{"page":1,"page_size":10,"total_rows":1}}`
	return []byte(res), nil
}
