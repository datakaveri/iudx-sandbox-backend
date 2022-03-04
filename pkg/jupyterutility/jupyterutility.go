package jupyterutility

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/httputility"
)

type JupyterUtility struct {
	HttpClient *httputility.HttpClient
}

type DeleteServerResponse struct {
	Message string `json:"message"`
}

type StopServerResponse struct {
	Message string `json:"message"`
}

type RestartServerResponse struct {
	Message string `json:"message"`
}

func Get() (*JupyterUtility, error) {
	return &JupyterUtility{
		HttpClient: httputility.Get(),
	}, nil
}

func (client *JupyterUtility) StopServer(app *application.Application, username, spawnerName string) (*StopServerResponse, error) {
	endpoint := fmt.Sprintf("%s/users/%s/servers/%s",
		app.Cfg.GetJupyterHubApi(), username, spawnerName)

	token := fmt.Sprintf("Bearer %s", app.Cfg.GetJupyterHubApiToken())
	res, err := client.HttpClient.SendRequest(endpoint, "DELETE", nil, token)

	if err != nil {
		log.Fatalf("Error stopping jupyter notebook %+v", err)
		return nil, err
	}

	stopServerResponse := &StopServerResponse{}
	json.Unmarshal(res, stopServerResponse)

	fmt.Println(stopServerResponse)

	return stopServerResponse, nil
}

func (client *JupyterUtility) DeleteServer(app *application.Application, username, spawnerName string) (*DeleteServerResponse, error) {
	endpoint := fmt.Sprintf("%s/users/%s/servers/%s",
		app.Cfg.GetJupyterHubApi(), username, spawnerName)

	fmt.Println(endpoint)

	token := fmt.Sprintf("Bearer %s", app.Cfg.GetJupyterHubApiToken())
	data := []byte(`{"remove": true }`)
	res, err := client.HttpClient.SendRequest(endpoint, "DELETE", bytes.NewBuffer(data), token)

	if err != nil {
		log.Fatalf("Error stopping jupyter notebook %+v", err)
		return nil, err
	}

	deleteServerResponse := &DeleteServerResponse{}
	json.Unmarshal(res, deleteServerResponse)

	fmt.Println(deleteServerResponse)

	return deleteServerResponse, nil
}

func (client *JupyterUtility) RestartServer(app *application.Application, username, spawnerName string) (*RestartServerResponse, string, error) {
	endpoint := fmt.Sprintf("%s/users/%s/servers/%s",
		app.Cfg.GetJupyterHubApi(), username, spawnerName)

	token := fmt.Sprintf("Bearer %s", app.Cfg.GetJupyterHubApiToken())
	res, err := client.HttpClient.SendRequest(endpoint, "POST", nil, token)

	if err != nil {
		log.Fatalf("Error restarting jupyter notebook %+v", err)
		return nil, "", err
	}

	restartServerResponse := &RestartServerResponse{}
	json.Unmarshal(res, restartServerResponse)

	return restartServerResponse, endpoint, nil
}
