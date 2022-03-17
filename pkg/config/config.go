package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	environment               string
	dbUser                    string
	dbPass                    string
	dbHost                    string
	dbPort                    string
	dbName                    string
	apiPort                   string
	binderHubApi              string
	jupyterHubApi             string
	jupyterHubApiToken        string
	keycloakPublicCertificate string
}

func Get() *Config {
	conf := &Config{}

	flag.StringVar(&conf.environment, "environment", os.Getenv("GO_ENV"), "Hosted environment")
	flag.StringVar(&conf.dbUser, "dbuser", os.Getenv("POSTGRES_USER"), "DB User Name")
	flag.StringVar(&conf.dbPass, "dbpass", os.Getenv("POSTGRES_PASSWORD"), "DB Password")
	flag.StringVar(&conf.dbHost, "dbhost", os.Getenv("POSTGRES_HOST"), "DB Host")
	flag.StringVar(&conf.dbPort, "dbport", os.Getenv("POSTGRES_PORT"), "DB Port")
	flag.StringVar(&conf.dbName, "dbname", os.Getenv("POSTGRES_DB"), "DB Name")
	flag.StringVar(&conf.apiPort, "apiPort", os.Getenv("API_PORT"), "API Port")
	flag.StringVar(&conf.binderHubApi, "binderHubApi", os.Getenv("BINDERHUB_API"), "Binderhub Notebook Build API")
	flag.StringVar(&conf.jupyterHubApi, "jupyterHubApi", os.Getenv("JUPYTERHUB_API"), "Jupyterhub API Host")
	flag.StringVar(&conf.jupyterHubApiToken, "jupyterHubApiToken", os.Getenv("JUPYTERHUB_API_TOKEN"), "Jupyterhub API Token")
	flag.StringVar(&conf.keycloakPublicCertificate, "keycloakPublicCertificate", os.Getenv("KEYCLOAK_PUBLIC_KEY"), "Keycloak Certificate Value excluding ")

	flag.Parse()

	return conf
}

func (c *Config) GetDBConnStr() string {
	return c.getDBConnStr(c.dbHost, c.dbName)
}

func (c *Config) getDBConnStr(dbhost, dbname string) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.dbUser,
		c.dbPass,
		dbhost,
		c.dbPort,
		dbname,
	)
}

func (c *Config) GetAPIPort() string {
	return ":" + c.apiPort
}

func (c *Config) GetBinderNotebookBuildApi(repoName string) string {
	return fmt.Sprintf(
		"%s/build/gh/datakaveri/%s/HEAD",
		c.binderHubApi,
		repoName,
		// token,
	)
}

func (c *Config) GetJupyterHubApi() string {
	return c.jupyterHubApi
}

func (c *Config) GetJupyterHubApiToken() string {
	return c.jupyterHubApiToken
}

func (c *Config) GetEnvironment() string {
	return c.environment
}

func (c *Config) GetKeycloakPublicCertificate() string {
	return "-----BEGIN CERTIFICATE-----\n" +
		c.keycloakPublicCertificate +
		"\n-----END CERTIFICATE-----"
}
