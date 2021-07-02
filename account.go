package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
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
	if account.bearer == "" {
		return nil, errors.New("Account is not authenticated!")
	}
	req.Header.Add("Authorization", "Bearer "+account.bearer)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

type MojangAccount struct {
	email             string
	password          string
	securityQuestions []SqAnswer
	securityAnswers   []string
	bearer            string
	uuid              string
	username          string
}

type AuthenticateReqResp struct {
	User struct {
		Properties []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"properties"`
		Username string `json:"username"`
		ID       string `json:"id"`
	} `json:"user"`
	Accesstoken string `json:"accessToken"`
	Clienttoken string `json:"clientToken"`
}

func (account *MojangAccount) authenticate() error {
	payload := fmt.Sprintf(`{
    "agent": {                              
        "name": "Minecraft",                
        "version": 1                        
    },
    "username": "%s",      
    "password": "%s",
	"requestUser": true
}`, account.email, account.password)

	u := bytes.NewReader([]byte(payload))
	request, err := http.NewRequest("POST", "https://authserver.mojang.com/authenticate", u)
	request.Header.Set("Content-Type", "application/json")

	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(request)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 300 {
		var AccountInfo AuthenticateReqResp
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		json.Unmarshal(b, &AccountInfo)

		account.bearer = AccountInfo.Accesstoken
		account.username = AccountInfo.User.Username
		account.uuid = AccountInfo.User.ID
		account.bearer = AccountInfo.Accesstoken

		return nil

	} else if resp.StatusCode == 403 {
		return errors.New("Invalid email or password!")
	}
	return errors.New("Reached end of authenticate function! Shouldn't be possible. most likely 'failed to auth' status code changed.")
}

type SqAnswer struct {
	Answer struct {
		ID int `json:"id"`
	} `json:"answer"`
	Question struct {
		ID       int    `json:"id"`
		Question string `json:"question"`
	} `json:"question"`
}

func (account *MojangAccount) loadSecurityQuestions() error {
	req, err := account.authenticatedReq("GET", "https://api.mojang.com/user/security/challenges", nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return errors.New(fmt.Sprintf("Got status %v when requesting security questions!", resp.Status))
	}

	defer resp.Body.Close()

	var sqAnswers []SqAnswer

	respBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	err = json.Unmarshal(respBytes, &sqAnswers)
	if err != nil {
		return err
	}

	account.securityQuestions = sqAnswers

	return nil
}

func (account *MojangAccount) needToAnswer() (bool, error) {
	req, err := account.authenticatedReq("GET", "https://api.mojang.com/user/security/location", nil)
	if err != nil {
		return false, err
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return true, err
	}

	if resp.StatusCode == 204 {
		return false, nil
	}
	if resp.StatusCode == 403 {
		return true, nil
	}
	return true, errors.New(fmt.Sprintf("Status of %v in needToAnswer not expected!", resp.Status))
}

type submitPostJson struct {
	ID     int    `json:"id"`
	Answer string `json:"answer"`
}

func (account *MojangAccount) submitAnswers() error {
	if len(account.securityAnswers) != 3 {
		return errors.New("Not enough security question answers provided!")
	}
	if len(account.securityQuestions) != 3 {
		return errors.New("Security questions not properly loaded!")
	}
	var jsonContent []submitPostJson
	for i, sq := range account.securityQuestions {
		jsonContent = append(jsonContent, submitPostJson{ID: sq.Answer.ID, Answer: account.securityAnswers[i]})
	}
	jsonStr, err := json.Marshal(jsonContent)
	if err != nil {
		return err
	}
	req, err := account.authenticatedReq("POST", "https://api.mojang.com/user/security/location", bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if resp.StatusCode == 204 {
		return nil
	}
	if resp.StatusCode == 403 {
		return errors.New("At least one security question answer was incorrect!")
	}
	return errors.New(fmt.Sprintf("Got status %v on post request for sqs", resp.Status))
}

func (account *MojangAccount) MojangAuthenticate() error {
	err := account.authenticate()
	if err != nil {
		return err
	}
	err = account.loadSecurityQuestions()

	if err != nil {
		return err
	}

	if len(account.securityQuestions) == 0 {
		return nil
	}

	answerNeeded, err := account.needToAnswer()
	if err != nil {
		return err
	}

	if !answerNeeded {
		return nil
	}

	err = account.submitAnswers()
	if err != nil {
		return err
	}

	return nil
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
