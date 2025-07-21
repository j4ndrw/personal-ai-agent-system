package agent

import (
	"io"
	"net/http"
)

func GetAgents() (*Agents, error) {
	req, err := http.NewRequest(http.MethodGet, GetAgentsEndpoint, nil)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	agents := Agents{}
	err = agents.ParseAgents(&body)
	if err != nil {
		return nil, err
	}

	return &agents, nil
}
