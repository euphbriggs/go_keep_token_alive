package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
)

// KeepAliveOptions contains the information necessary to reach the API
type KeepAliveOptions struct {
	APIHost   string
	LoginPath string
	Port      string
	PingPath  string
}

// LoginInfo contains the email and password to connect with
type LoginInfo struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse contains the token from the API
type LoginResponse struct {
	Token string `json:"token"`
}

// RefreshResponse contains the response from the API after an attempt to refresh a session token
type RefreshResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func main() {
	// Get the required flags
	var apiHost = flag.String("apiHost", "localhost", "API Hostname to connect to")
	var apiPort = flag.String("apiPort", "4000", "Port the API is listening on")
	var loginPath = flag.String("loginPath", "/api/login", "API endpoint to login to")
	var pingPath = flag.String("pingPath", "/api/refresh_session", "API endpoint to hit when refreshing the session")

	var loginEmail = flag.String("loginEmail", "", "Email to login with")
	var loginPassword = flag.String("loginPassword", "", "Password to login with")
	var existingToken = flag.String("token", "", "Token to use to keep alive")

	flag.Parse()

	if *loginEmail == "" || *loginPassword == "" {
		fmt.Println("loginEmail and loginPassword are required")
		return
	}

	options := KeepAliveOptions{
		APIHost:   *apiHost,
		LoginPath: *loginPath,
		Port:      *apiPort,
		PingPath:  *pingPath,
	}

	loginInfo := LoginInfo{
		Email:    *loginEmail,
		Password: *loginPassword,
	}

	if *existingToken == "" {
		token, err := login(options, loginInfo)
		if err != nil {
			panic(err)
		}

		fmt.Println(token)
	} else {
		token, err := refreshToken(options, loginInfo, *existingToken)
		if err != nil {
			panic(err)
		}

		fmt.Println(token)
	}
}

func refreshToken(options KeepAliveOptions, loginInfo LoginInfo, token string) (string, error) {
	url := fmt.Sprintf("http://%s:%s/%s", options.APIHost, options.Port, options.PingPath)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var refreshResponse RefreshResponse
	err = json.Unmarshal(body, &refreshResponse)
	if err != nil {
		return "", err
	}

	// The session wasn't successfully refresh, so get a new token
	if !refreshResponse.Success {
		return login(options, loginInfo)
	}

	// The session was refreshed successfully, return the existing token
	return token, nil
}

func login(options KeepAliveOptions, loginInfo LoginInfo) (string, error) {
	url := fmt.Sprintf("http://%s:%s/%s", options.APIHost, options.Port, options.LoginPath)

	loginBytes, err := json.Marshal(loginInfo)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(loginBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var loginResponse LoginResponse
	err = json.Unmarshal(body, &loginResponse)
	if err != nil {
		return "", err
	}

	return loginResponse.Token, nil
}
