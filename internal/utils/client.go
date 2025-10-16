package utils

import (
	"bufio"
	"fmt"
	"net/http"
)

var url string = "https://app.addy.io"
var ver string = "v1"
var api string = url + "/api" + "/" + ver

func NewClient() (*http.Client, error) {
	client := &http.Client{
		// CheckRedirect: redirectPolicyFunc,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	switch s := resp.StatusCode; {
	case 200 <= s && s <= 299:
		return client, nil
	case 300 <= s && s <= 399:
		return client, nil
	case 400 <= s && s <= 499:
		return nil, fmt.Errorf("Request Side Error status code: %d", s)
	case 500 <= s && s <= 599:
		return nil, fmt.Errorf("Server Side Error status code: %d", s)
	default:
		return nil, fmt.Errorf("unexpected status code: %d", s)
	}
}

func Curl(client *http.Client, endpoint string, method string, token string) {
	var bearer string = "Bearer " + token

	req, err := http.NewRequest("GET", api+"/"+endpoint, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Authorization:", bearer)
	req.Header.Add("Content-Type:", "application/json")
	req.Header.Add("X-Requested-With:", "XMLHttpRequest")

	resp, err := client.Do(req)

	defer resp.Body.Close()

	// TODO: change to float the error up w/ context
	fmt.Println("Response status:", resp.Status)

	scanner := bufio.NewScanner(resp.Body)
	for i := 0; scanner.Scan() && i < 5; i++ {
		fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
