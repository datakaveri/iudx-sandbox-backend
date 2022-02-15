package httputility

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type HttpClient struct {
	httpClient *http.Client
}

const (
	RequestTimeout int = 30
)

func Get() *HttpClient {
	client := &http.Client{Timeout: time.Duration(RequestTimeout) * time.Second}
	return &HttpClient{
		httpClient: client,
	}
}

func (client *HttpClient) SendRequest(endpoint string, method string, data io.Reader, token string) ([]byte, error) {
	req, err := http.NewRequest(method, endpoint, data)

	if token != "" {
		req.Header.Set("Authorization", token)
	}

	if err != nil {
		log.Fatalf("Error Occured %+v", err)
	}

	response, err := client.httpClient.Do(req)

	if err != nil {
		log.Fatalf("Error sending request to API endpoint %+v", err)
		return nil, err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Fatalf("Couldn't parse response body %+v", err)
	}

	return body, nil
}
