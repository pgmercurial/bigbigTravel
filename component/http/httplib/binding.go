package httplib

import (
	"net/http"
	"encoding/json"
	"io/ioutil"
)

type jsonBinding struct{}

func (jsonBinding) Name() string {
	return "json"
}

func (jsonBinding) Bind(req *http.Request, obj interface{}) ([]byte, error) {
	jsonString,err := ioutil.ReadAll(req.Body)
	if err != nil {
		return []byte{}, err
	}
	json.Unmarshal(jsonString, obj)
	if err := json.Unmarshal(jsonString, obj); err != nil {
		return []byte{}, err
	}

	return jsonString, nil
}