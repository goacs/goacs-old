package lib

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func GetFileUrl(path string, request *http.Request) (string, error) {
	libenv := Env{}

	fileDir := libenv.Get("FILESTORE_PATH", "./storage")
	//TODO: Security check
	absPath, _ := filepath.Abs(fileDir)
	if _, err := os.Stat(absPath + "/" + path); os.IsNotExist(err) {
		fmt.Printf("File does not exist\n")
		return "", err
	}

	urlStruct := url.URL{
		Scheme: "http",
		Host:   request.Host,
		Path:   "file/" + path,
	}

	return urlStruct.String(), nil
}
