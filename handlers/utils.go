package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"golang.org/x/oauth2"
)

func getUserData(conf oauth2.Config, token *oauth2.Token) (map[string]interface{}, error) {
	client := conf.Client(context.TODO(), token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, fmt.Errorf("issue instantiating oauth2 client: %v", err)
	}
	defer resp.Body.Close()

	userData := make(map[string]interface{})

	data, _ := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(data, &userData)
	if err != nil {
		return nil, fmt.Errorf("issue unmarshalling data from google: %v", err)
	}

	return userData, nil
}
