package utils

import (
	"bufio"
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var url string = "https://app.addy.io"
var ver string = "v1"
var api string = url + "/api" + "/" + ver

func NewClient(ctx context.Context) (*http.Client, error) {
	tflog.Info(ctx, "Creating http.Client")
	client := &http.Client{
		// CheckRedirect: redirectPolicyFunc,
	}
	tflog.Trace(ctx, "http.Client Created: ") // TODO: pretty print out client object

	tflog.Trace(ctx, "Sending request to "+url+" as test")
	resp, err := client.Get(url)
	if err != nil {
		tflog.Error(ctx, "Error on client init: [GET] "+url)
		// TODO: tflog 'err'
		return nil, err
	}

	switch s := resp.StatusCode; {
	case 200 <= s && s <= 299:
		tflog.Debug(ctx, "Response status: "+resp.Status)
		return client, nil
	case 300 <= s && s <= 399:
		tflog.Warn(ctx, "Response status: "+resp.Status)
		return client, nil
	case 400 <= s && s <= 499:
		tflog.Error(ctx, "Response status: "+resp.Status)
		return nil, fmt.Errorf("Request Side Error status code: %d", s)
	case 500 <= s && s <= 599:
		tflog.Error(ctx, "Response status: "+resp.Status)
		return nil, fmt.Errorf("Server Side Error status code: %d", s)
	default:
		tflog.Error(ctx, "Response status: "+resp.Status)
		return nil, fmt.Errorf("unexpected status code: %d", s)
	}
}

func Curl(ctx context.Context, client *http.Client, endpoint string, method string, token string) {
	var bearer string = "Bearer " + token

	req, err := http.NewRequest("GET", api+"/"+endpoint, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Authorization:", bearer)
	req.Header.Add("Content-Type:", "application/json")
	req.Header.Add("X-Requested-With:", "XMLHttpRequest")

	tflog.Trace(ctx, "Sending request: "+endpoint) // TODO: pretty print the 'req' object

	resp, err := client.Do(req)

	defer resp.Body.Close()

	tflog.Debug(ctx, "Response status: "+resp.Status)

	scanner := bufio.NewScanner(resp.Body)
	for i := 0; scanner.Scan() && i < 5; i++ {
		fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
