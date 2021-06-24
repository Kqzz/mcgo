package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

func (account *MojangAccount) authenticatedReq(method string, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", account.bearer)

	return req, nil
}

type MojangAccount struct {
	email    string
	password string
	bearer   string
	uuid     string
	username string
}

func (account *MojangAccount) authenticate() (string, error) {
	return "success", nil
}

type nameChangeInfoResponse struct {
	Changedat         time.Time `json:"changedAt"`
	Createdat         time.Time `json:"createdAt"`
	Namechangeallowed bool      `json:"nameChangeAllowed"`
}

func (account *MojangAccount) nameChangeInfo() (nameChangeInfoResponse, error) {
	client := &http.Client{}
	req, err := account.authenticatedReq("GET", "https://api.minecraftservices.com/minecraft/profile/namechange", nil)

	if err != nil {
		return nameChangeInfoResponse{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nameChangeInfoResponse{}, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nameChangeInfoResponse{}, err
	}

	var parsedNameChangeInfo nameChangeInfoResponse

	err = json.Unmarshal(respBody, &parsedNameChangeInfo)

	if err != nil {
		return nameChangeInfoResponse{}, err
	}

	return parsedNameChangeInfo, nil
}

func (account *MojangAccount) changeName(username string, changeTime time.Time) (string, error) {
	if changeTime.After(time.Now()) {
		// wait until changeTime
		time.Sleep(time.Until(changeTime))
	}
	return fmt.Sprintf("Changed name of %v to %v", account.email, username), nil
}
