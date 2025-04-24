package main

import (
	"log"
	"net/http"
	"net/url"
)

// Limit represents the maximum number of integrations we are going to fetch per page.
const Limit = 100

// MaximumConcurrency sets the maximum number of concurrent workers that the application will launch.
const MaximumConcurrency = 0

// NameFilter is the filter that will be applied when querying for integrations.
const NameFilter = ""

// Password is the password to be used in the basic authentication header.
const Password = ""

// TypesFilter specifies which "integration type" filter to apply.
var TypesFilter = []string{"camel:google_chat", "camel:slack", "camel:teams", "webhook"}

// Username is the username to be used in the basic authentication header.
const Username = ""

func main() {
	integrationsGenericUrl, err := url.Parse("https://console.stage.redhat.com/api/integrations/v1.0/endpoints")
	if err != nil {
		log.Fatalf("unable to parse generic URL: %s", err)
	}

	httpClient, err := NewHttpClient(true)
	if err != nil {
		log.Fatalf("unable to create HTTP client: %s", err)
	}

	// The semaphore allows us to control the maximum number of concurrent goroutines that are doing work.
	//var semaphore = make(chan struct{}, 1)

	offset := 0
	for {
		var responseBody GetIntegrationsResponse
		err = MakeRequest(httpClient, http.MethodGet, integrationsGenericUrl.String(), Limit, offset, NameFilter, TypesFilter, http.StatusOK, &responseBody)
		if err != nil {
			log.Printf("[Limit: %d][offset: %d] Unable to get the list of integrations: %s\n", Limit, offset, err)
			offset = offset + Limit

			continue
		}

		log.Printf(`Fetched %d integrations with "%s" in the name`, len(responseBody.Integrations), NameFilter)
		//semaphore <- struct{}{}

		// Remove all the fetched integrations.
		for _, integration := range responseBody.Integrations {
			deleteUrl := integrationsGenericUrl.JoinPath("/", integration.Id)

			if err = MakeRequest(httpClient, http.MethodDelete, deleteUrl.String(), 0, 0, "", nil, http.StatusNoContent, nil); err != nil {
				log.Printf("[integration_id: %s] Unable to delete integration: %s\n", integration.Id, err)
				continue
			}

			log.Printf("[integration_id: %s][integration_name: %s] Integration deleted\n", integration.Id, integration.Name)
		}

		//<-semaphore

		// Update the offset so that we can request the next page.
		offset = offset + len(responseBody.Integrations)

		if len(responseBody.Integrations) != Limit {
			break
		}
	}
}
