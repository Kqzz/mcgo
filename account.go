package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
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

type nameChangeReturn struct {
	account     MojangAccount
	username    string
	changedName bool
	statusCode  int
	sendTime    time.Time
	receiveTime time.Time
}

func (account *MojangAccount) changeName(username string, changeTime time.Time) (nameChangeReturn, error) {

	headers := make(http.Header)
	headers.Add("Authorization", "Bearer "+account.bearer)
	payload, err := generatePayload("PUT", fmt.Sprintf("https://api.minecraftservices.com/minecraft/profile/name/%s", username), headers)

	recvd := make([]byte, 12)

	if err != nil {
		return nameChangeReturn{
			account:     MojangAccount{},
			username:    username,
			changedName: false,
			statusCode:  0,
			sendTime:    time.Time{},
			receiveTime: time.Time{},
		}, err
	}

	if changeTime.After(time.Now()) {
		// wait until 20s before nc
		time.Sleep(time.Until(changeTime) - time.Second*20)
	}

	conn, err := tls.Dial("tcp", "api.minecraftservices.com"+":443", nil)
	conn.Write([]byte(payload))
	sendTime := time.Now()
	if err != nil {
		return nameChangeReturn{
			account:     MojangAccount{},
			username:    username,
			changedName: false,
			statusCode:  0,
			sendTime:    sendTime,
			receiveTime: time.Time{},
		}, err
	}

	conn.Write([]byte("\r\n"))

	conn.Read(recvd)
	recvTime := time.Now()
	status, err := strconv.Atoi(string(recvd[9:12]))

	toRet := nameChangeReturn{
		account:     *account,
		username:    username,
		changedName: status < 300,
		statusCode:  status,
		sendTime:    sendTime,
		receiveTime: recvTime,
	}
	return toRet, nil
}
