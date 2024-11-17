package places

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

const BASE_MAPBOX_API = "https://api.mapbox.com"

type APIEndpoint string

const (
	Autosuggest APIEndpoint = "/search/searchbox/v1/suggest"
)

type MapboxApi struct {
	Method       string
	Endpoint     string
	QueryParam   string
	Payload      string
	AccessToken  string
	SessionToken string
}

func GetAPIUrl(data MapboxApi) string {
	if data.Method == "GET" {
		return fmt.Sprintf("%s%s?q=%s&access_token=%s&session_token=%s", BASE_MAPBOX_API, data.Endpoint, data.QueryParam, data.AccessToken, data.SessionToken)
	}
	return fmt.Sprintf("%s%s?access_token=%s&session_token=%s", BASE_MAPBOX_API, data.Endpoint, data.AccessToken, data.SessionToken)

}

func make_http_request(data MapboxApi) string {
	url := GetAPIUrl(data)
	resp, err := http.Get(url)
	if err != nil {
		// handle error
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	return string(body)
}

// func (c *Client) callAPI(ctx context.Context, r *request, opts ...RequestOption) (data []byte, err error) {
// 	err = c.parseRequest(r, opts...)
// 	if err != nil {
// 		return []byte{}, err
// 	}
// 	req, err := http.NewRequest(r.method, r.fullURL, r.body)
// 	if err != nil {
// 		return []byte{}, err
// 	}
// 	req = req.WithContext(ctx)
// 	req.Header = r.header
// 	c.debug("request: %#v\n", req)
// 	f := c.do
// 	if f == nil {
// 		f = c.HTTPClient.Do
// 	}
// 	res, err := f(req)
// 	if err != nil {
// 		return []byte{}, err
// 	}
// 	data, err = io.ReadAll(res.Body)
// 	if err != nil {
// 		return []byte{}, err
// 	}
// 	defer func() {
// 		cerr := res.Body.Close()
// 		// Only overwrite the returned error if the original error was nil and an
// 		// error occurred while closing the body.
// 		if err == nil && cerr != nil {
// 			err = cerr
// 		}
// 	}()
// 	c.debug("response: %#v\n", res)
// 	c.debug("response body: %s\n", string(data))
// 	c.debug("response status code: %d\n", res.StatusCode)
//
// 	if res.StatusCode >= http.StatusBadRequest {
// 		apiErr := new(common.APIError)
// 		e := json.Unmarshal(data, apiErr)
// 		if e != nil {
// 			c.debug("failed to unmarshal json: %s\n", e)
// 		}
// 		if !apiErr.IsValid() {
// 			apiErr.Response = data
// 		}
// 		return nil, apiErr
// 	}
// 	return data, nil
// }
