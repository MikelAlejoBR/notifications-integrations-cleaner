package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// proxyURL is the URL for the proxy in the stage environment.
const proxyURL = ""

func NewHttpClient(isStage bool) (*http.Client, error) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	if isStage {
		proxyURL, err := url.Parse(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("error parsing the proxy URL: %w", err)
		}

		httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	return httpClient, nil
}

func MakeRequest(httpClient *http.Client, httpMethod string, targetUrl string, limit, offset int, nameFilter string, typesFilter []string, expectedStatusCode int, marshalTo any) error {
	request, err := http.NewRequest(httpMethod, targetUrl, nil)
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	request.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(Username+":"+Password)))

	queryParameters := request.URL.Query()
	if limit != 0 {
		queryParameters.Add("limit", strconv.Itoa(limit))
	}

	if offset != 0 {
		queryParameters.Add("offset", strconv.Itoa(offset))
	}

	if nameFilter != "" {
		queryParameters.Add("name", nameFilter)
	}

	if typesFilter != nil && len(typesFilter) > 0 {
		for _, filter := range typesFilter {
			queryParameters.Add("type", filter)
		}
	}

	request.URL.RawQuery = queryParameters.Encode()

	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("error when making request: %w", err)
	}
	defer response.Body.Close()

	rawBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error while reading the response body: %w", err)
	}

	if response.StatusCode != expectedStatusCode {
		return fmt.Errorf("unexpected status code received. Want %d, got %d", expectedStatusCode, response.StatusCode)
	}

	if marshalTo != nil {
		if len(rawBody) > 0 {
			if err := json.Unmarshal(rawBody, &marshalTo); err != nil {
				return fmt.Errorf("unable to unmarsal response body: %w", err)
			}
		} else {
			log.Println("Received response body is zero. Not marshaling response body.")
		}
	}

	return nil
}
