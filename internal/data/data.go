package data

import (
	"net/http"
)

type DataSourceData struct {
	Client *http.Client
	ApiKey string
}