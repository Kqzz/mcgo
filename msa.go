package mcgo

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

func authServerHandler(c chan result, state string, msConfig *oauth2.Config) {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Println("Got Req")
		// verify that it is the correct state
		if state == request.URL.Query().Get("state") {
			// obtain the code from the URL and make the exchange to get the token.
			token, err := msConfig.Exchange(context.Background(), request.URL.Query().Get("code"))

			if err != nil {
				// Pass error along to the channel
				c <- result{
					Value: nil,
					Err:   err,
				}
			}
			// pass it along to the channel to actually use the token
			c <- result{
				Value: token,
				Err:   nil,
			}
		}
	})

	fmt.Println("Listening on port 2445!")
	fmt.Println(http.ListenAndServe(":2445", nil))

}

type result struct {
	Value *oauth2.Token
	Err   error
}

type xBLSignInBody struct {
	Properties struct {
		Authmethod string `json:"AuthMethod"`
		Sitename   string `json:"SiteName"`
		Rpsticket  string `json:"RpsTicket"`
	} `json:"Properties"`
	Relyingparty string `json:"RelyingParty"`
	Tokentype    string `json:"TokenType"`
}

type XBLSignInResp struct {
	Issueinstant  time.Time `json:"IssueInstant"`
	Notafter      time.Time `json:"NotAfter"`
	Token         string    `json:"Token"`
	Displayclaims struct {
		Xui []struct {
			Uhs string `json:"uhs"`
		} `json:"xui"`
	} `json:"DisplayClaims"`
}

type xSTSPostBody struct {
	Properties struct {
		Sandboxid  string   `json:"SandboxId"`
		Usertokens []string `json:"UserTokens"`
	} `json:"Properties"`
	Relyingparty string `json:"RelyingParty"`
	Tokentype    string `json:"TokenType"`
}

type xSTSAuthorizeResponse struct {
	Issueinstant  time.Time `json:"IssueInstant"`
	Notafter      time.Time `json:"NotAfter"`
	Token         string    `json:"Token"`
	Displayclaims struct {
		Xui []struct {
			Uhs string `json:"uhs"`
		} `json:"xui"`
	} `json:"DisplayClaims"`
}

type xSTSAuthorizeResponseFail struct {
	Identity string `json:"Identity"`
	Xerr     int64  `json:"XErr"`
	Message  string `json:"Message"`
	Redirect string `json:"Redirect"`
}

type msGetMojangbearerBody struct {
	Identitytoken       string `json:"identityToken"`
	Ensurelegacyenabled bool   `json:"ensureLegacyEnabled"`
}

type msGetMojangBearerResponse struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
	Foci         string `json:"foci"`
}

func (account *MCaccount) MicrosoftAuthenticate(clientID, ClientSecret string) error {
	rand.Seed(time.Now().UnixNano())
	msConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: ClientSecret,
		Endpoint:     microsoft.LiveConnectEndpoint,
		RedirectURL:  "http://localhost:2445",
		Scopes:       []string{"XboxLive.signin", "XboxLive.offline_access"},
	}
	state := "MCS" + strconv.Itoa(int(rand.Int63()))

	fmt.Println(msConfig.AuthCodeURL(state))

	x := make(chan result)

	go authServerHandler(x, state, msConfig)

	tokenResult := <-x

	if tokenResult.Err != nil {
		return tokenResult.Err
	}
	token := tokenResult.Value
	client := msConfig.Client(context.Background(), tokenResult.Value)
	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			Renegotiation:      tls.RenegotiateOnceAsClient,
			InsecureSkipVerify: true,
		},
	}

	data := xBLSignInBody{
		Properties: struct {
			Authmethod string "json:\"AuthMethod\""
			Sitename   string "json:\"SiteName\""
			Rpsticket  string "json:\"RpsTicket\""
		}{
			Authmethod: "RPS",
			Sitename:   "user.auth.xboxlive.com",
			Rpsticket:  "d=" + token.AccessToken,
		},
		Relyingparty: "http://auth.xboxlive.com",
		Tokentype:    "JWT",
	}

	encodedBody, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", "https://user.auth.xboxlive.com/user/authenticate", bytes.NewReader(encodedBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-xbl-contract-version", "1")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	var respBody XBLSignInResp

	json.Unmarshal(respBodyBytes, &respBody)

	uhs := respBody.Displayclaims.Xui[0].Uhs
	XBLToken := respBody.Token

	xstsBody := xSTSPostBody{
		Properties: struct {
			Sandboxid  string   "json:\"SandboxId\""
			Usertokens []string "json:\"UserTokens\""
		}{
			Sandboxid: "RETAIL",
			Usertokens: []string{
				XBLToken,
			},
		},
		Relyingparty: "rp://api.minecraftservices.com/",
		Tokentype:    "JWT",
	}

	encodedXstsBody, err := json.Marshal(xstsBody)
	if err != nil {
		return err
	}
	req, err = http.NewRequest("POST", "https://xsts.auth.xboxlive.com/xsts/authorize", bytes.NewReader(encodedXstsBody))
	if err != nil {
		return err
	}

	resp, err = client.Do(req)

	if err != nil {
		return err
	}

	respBodyBytes, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if resp.StatusCode == 401 {
		var authorizeXstsFail xSTSAuthorizeResponseFail
		json.Unmarshal(respBodyBytes, &authorizeXstsFail)
		switch authorizeXstsFail.Xerr {
		case 2148916238:
			{
				return errors.New("microsoft account belongs to someone under 18! add to family for this to work")
			}
		case 2148916233:
			{
				return errors.New("you have no xbox account! Sign up for one to continue")
			}
		default:
			{
				return fmt.Errorf("got error code %v when trying to authorize XSTS token", authorizeXstsFail.Xerr)
			}
		}
	}

	var xstsAuthorizeResp xSTSAuthorizeResponse
	json.Unmarshal(respBodyBytes, &xstsAuthorizeResp)

	xstsToken := xstsAuthorizeResp.Token

	mojangBearerBody := msGetMojangbearerBody{
		Identitytoken:       "XBL3.0 x=" + uhs + ";" + xstsToken,
		Ensurelegacyenabled: true,
	}

	mojangBearerBodyEncoded, err := json.Marshal(mojangBearerBody)

	if err != nil {
		return err
	}

	req, err = http.NewRequest("POST", "https://api.minecraftservices.com/authentication/login_with_xbox", bytes.NewReader(mojangBearerBodyEncoded))

	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return err
	}

	resp, err = client.Do(req)
	if err != nil {
		return err
	}

	mcBearerResponseBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	fmt.Println(string(mcBearerResponseBytes))

	var mcBearerResp msGetMojangBearerResponse

	json.Unmarshal(mcBearerResponseBytes, &mcBearerResp)

	account.Bearer = mcBearerResp.AccessToken

	return nil
}

// Thanks to https://mojang-api-docs.netlify.app/ for the docs
// and https://gist.github.com/rbrick/be8ed86864fc5d77aa6c979053cfc892 for a great example of msa implemented in go!
