package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Droptime struct {
	droptime time.Time
	username string
}

func getDroptimeKqzzAPI(username string) (Droptime, error) {
	url := fmt.Sprintf("https://droptime-o7u637bu7a-uc.a.run.app/droptime/%v", username)
	resp, err := http.Get(url)
	if err != nil {
		return Droptime{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return Droptime{}, errors.New(fmt.Sprintf("%v is not dropping!", username))
	}
	body, err := ioutil.ReadAll(resp.Body)
	bodyStr := string(body)
	droptimeInt, err := strconv.ParseInt(bodyStr, 10, 64)
	return Droptime{
		droptime: time.Unix(droptimeInt, 0),
		username: username,
	}, nil
}

type TeunResponse struct {
	UNIX int64  `json:"UNIX"`
	UTC  string `json:"UTC"`
}

func getDroptimeTeunAPI(username string) (Droptime, error) {
	url := fmt.Sprintf("https://api.teun.lol/droptime/%v", username)
	resp, err := http.Get(url)
	if err != nil {
		return Droptime{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 300 {
		droptime := TeunResponse{
			UNIX: 0,
			UTC:  "",
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return Droptime{}, err
		}
		err = json.Unmarshal(body, &droptime)

		fmt.Println(droptime)

		return Droptime{
			time.Unix(droptime.UNIX, 0),
			username,
		}, nil

	}

	return Droptime{}, nil

}

func manualDroptime(droptime int64) time.Time {
	droptimeParsed := time.Unix(droptime, 0)
	return droptimeParsed
}

func nameAvailability(username string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.mojang.com/user/profile/agent/minecraft/name/%v", username))

	if err != nil {
		return "", err
	}

	if resp.StatusCode == 200 {
		return "claimed", nil
	}

	if resp.StatusCode == 400 {
		return "invalid", nil
	}

	if resp.StatusCode == 429 {
		return "", errors.New("Mojang API ratelimit reached!")
	}

	return "", errors.New(fmt.Sprintf("This should not be possible! | Got status %v on request for name availability", resp.StatusCode))
}

func generatePayload(method string, reqUrl string, headers http.Header) (string, error) {
	parsedUrl, err := url.Parse(reqUrl)
	if err != nil {
		return "", err
	}
	host := parsedUrl.Host
	path := parsedUrl.Path
	var headerString string
	for header, value := range headers {
		headerString += fmt.Sprintf("%s: %s\r\n", header, value[0])
	}
	payload := fmt.Sprintf("%s %s HTTP/1.1\r\nHost: %s\r\n%s\r\n", strings.ToUpper(method), path, host, headerString)
	return payload, nil
}
