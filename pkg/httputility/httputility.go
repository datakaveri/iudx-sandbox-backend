package httputility

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func httpClient() *http.Client {
	client := &http.Client{Timeout: 10 * time.Second}
	return client
}

func sendRequest(client *http.Client, endpoint string, method string, data io.Reader) []byte {
	req, err := http.NewRequest(method, endpoint, data)

	if err != nil {
		log.Fatalf("Error Occured %+v", err)
	}

	response, err := client.Do(req)

	if err != nil {
		log.Fatalf("Error sending request to API endpoint %+v", err)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Fatalf("Couldn't parse response body %+v", err)
	}

	return body
}
