package jupyterutility

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/iudx-sandbox-backend/pkg/application"
	"github.com/iudx-sandbox-backend/pkg/httputility"
)

type JupyterUtility struct {
	HttpClient *httputility.HttpClient
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

func (client *JupyterUtility) StopServer(app *application.Application, username string) (*StopServerResponse, error) {
	endpoint := fmt.Sprintf("%s/users/%s/server",
		app.Cfg.GetJupyterHubApi(), username)

	res, err := client.HttpClient.SendRequest(endpoint, "DELETE", nil)

	if err != nil {
		log.Fatalf("Error stopping jupyter notebook %+v", err)
		return nil, err
	}

	stopServerResponse := &StopServerResponse{}
	json.Unmarshal(res, stopServerResponse)

	return stopServerResponse, nil
}

func (client *JupyterUtility) RestartServer(app *application.Application, username string) (*RestartServerResponse, error) {
	endpoint := fmt.Sprintf("%s/users/%s/server",
		app.Cfg.GetJupyterHubApi(), username)

	res, err := client.HttpClient.SendRequest(endpoint, "POST", nil)

	if err != nil {
		log.Fatalf("Error restarting jupyter notebook %+v", err)
		return nil, err
	}

	restartServerResponse := &RestartServerResponse{}
	json.Unmarshal(res, restartServerResponse)

	return restartServerResponse, nil
}
