package main

type GetIntegrationsResponse struct {
	Integrations []Integration `json:"data"`
}

type Integration struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
