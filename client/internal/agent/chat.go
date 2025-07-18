package agent

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type OnChunk func(chunk AgentChunk)

func Chat(
	prompt string,
	onChunk OnChunk,
	onDone func(),
) error {
	agentChunk := AgentChunk{}

	bodyData, err := json.Marshal(map[string]string{"prompt": prompt})
	if err != nil {
		return err
	}

	resp, err := http.Post(
		"http://localhost:6969/api/chat",
		"application/json",
		bytes.NewBuffer(bodyData),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	for {
		buf := make([]byte, 1024)
		_, err := resp.Body.Read(buf)
		buf = bytes.Trim(buf, "\x00")

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		err = agentChunk.ParseAgentChunk(&buf)
		if err != nil {
			return err
		}
		onChunk(agentChunk)
	}
	onDone()
	return nil
}
