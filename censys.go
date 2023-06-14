package censyssearchgo

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type SearchType int

const (
	Host SearchType = iota
)

var API_ID string
var SECRET string

// callback function
type HitsCallback func(hits Hits)

func queryFromCensys(url string) (string, error) {

	method := "GET"

	//避免x509: certificate signed by unknown authority
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/json")

	req.SetBasicAuth(API_ID, SECRET)

	res, err := client.Do(req)
	if err != nil {

		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func dealWithHostSearchResponse(url string, responseJson string, callback HitsCallback) error {
	var result HostSearchResult
	err := json.Unmarshal([]byte(responseJson), &result)
	if err != nil {
		return fmt.Errorf("error when unmarshal json result: %s", err)
	}

	if result.Code != 200 && result.Status != "OK" {
		return fmt.Errorf("error when search host: %s", result.Status)
	}

	for _, hit := range result.Result.Hits {
		callback(hit)
	}

	if result.Result.Links.Next != "" {
		nextUrl := url + "&cursor=" + result.Result.Links.Next
		responseJson, err = queryFromCensys(nextUrl)
		if err != nil {
			return err
		}
		dealWithHostSearchResponse(url, responseJson, callback)
	}

	return nil
}

func searchHost(query string, includeVirtualHosts bool, callback HitsCallback) error {

	urlEncodedQuery := url.QueryEscape(query)

	url := "https://search.censys.io/api/v2/hosts/search?q=" + urlEncodedQuery +
		"&per_page=100&virtual_hosts=" + (map[bool]string{true: "EXCLUDE", false: "INCLUDE"})[includeVirtualHosts]

	//fmt.Println(url)

	responseJson, err := queryFromCensys(url)
	if err != nil {
		return err
	}

	return dealWithHostSearchResponse(url, responseJson, callback)
}

// searchType: The type of search to perform, currently only Host is supported
//
// query: The query to search for, example: services.http.response.body:"xxx"  , more : https://search.censys.io/search/definitions?resource=hosts
//
// includeVirtualHosts: Whether to include virtual hosts in the results
//
// HitsCallback: The callback function to call when a result is found
//
// api_id: The API ID for the Censys API, find your API at: https://search.censys.io/account/api
//
// api_secret: The secret for the Censys API
func Search(searchType SearchType, query string, includeVirtualHosts bool, callback HitsCallback, api_id string, api_secret string) error {
	API_ID = api_id
	SECRET = api_secret

	switch searchType {
	case Host:
		return searchHost(query, includeVirtualHosts, callback)
	default:
		return nil
	}
}
