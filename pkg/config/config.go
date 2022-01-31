package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	dbUser        string
	dbPass        string
	dbHost        string
	dbPort        string
	dbName        string
	apiPort       string
	binderHubApi  string
	jupyterHubApi string
}

func Get() *Config {
	conf := &Config{}

	flag.StringVar(&conf.dbUser, "dbuser", os.Getenv("POSTGRES_USER"), "DB User Name")
	flag.StringVar(&conf.dbPass, "dbpass", os.Getenv("POSTGRES_PASSWORD"), "DB Password")
	flag.StringVar(&conf.dbHost, "dbhost", os.Getenv("POSTGRES_HOST"), "DB Host")
	flag.StringVar(&conf.dbPort, "dbport", os.Getenv("POSTGRES_PORT"), "DB Port")
	flag.StringVar(&conf.dbName, "dbname", os.Getenv("POSTGRES_DB"), "DB Name")
	flag.StringVar(&conf.apiPort, "apiPort", os.Getenv("API_PORT"), "API Port")
	flag.StringVar(&conf.binderHubApi, "binderHubApi", os.Getenv("BINDERHUB_API"), "Binderhub Notebook Build API")
	flag.StringVar(&conf.jupyterHubApi, "jupyterHubApi", os.Getenv("JUPYTERHUB_API"), "Jupyterhub API Host")

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
		"%s/build/gh/swarup-e/%s/HEAD",
		c.binderHubApi,
		repoName,
		// token,
	)
}

func (c *Config) GetJupyterHubApi() string {
	return c.jupyterHubApi
}
