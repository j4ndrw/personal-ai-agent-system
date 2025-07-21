package agent

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"slices"
	"strings"
	"time"

	"net/http"

	"github.com/j4ndrw/personal-ai-agent-system/client/internal/async"
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/state"
)

type ReceiveStreamChunkMsg struct {
	AgentChunk AgentChunk
	Response   *http.Response
	Time       time.Time
}
type ReceiveStreamChunkTickMsg ReceiveStreamChunkMsg

func OpenStream(
	prompt string,
	endpoint string,
) (ReceiveStreamChunkMsg, error) {
	bodyData, err := json.Marshal(map[string]string{"prompt": prompt})
	if err != nil {
		return ReceiveStreamChunkMsg{}, err
	}

	http.DefaultClient.Timeout = 0
	resp, err := http.Post(
		endpoint,
		"application/json",
		bytes.NewBuffer(bodyData),
	)
	if err != nil {
		return ReceiveStreamChunkMsg{}, err
	}

	return ReceiveStreamChunkMsg{
		AgentChunk: AgentChunk{},
		Response:   resp,
		Time:       time.Now(),
	}, nil
}

func OpenAgenticManualStream(
	prompt string,
	agent string,
	endpoint string,
) (ReceiveStreamChunkMsg, error) {
	bodyData, err := json.Marshal(map[string]string{"prompt": prompt, "agent": agent})
	if err != nil {
		return ReceiveStreamChunkMsg{}, err
	}

	http.DefaultClient.Timeout = 0
	resp, err := http.Post(
		endpoint,
		"application/json",
		bytes.NewBuffer(bodyData),
	)
	if err != nil {
		return ReceiveStreamChunkMsg{}, err
	}

	return ReceiveStreamChunkMsg{
		AgentChunk: AgentChunk{},
		Response:   resp,
		Time:       time.Now(),
	}, nil
}

func ReadChunk(msg ReceiveStreamChunkMsg, rcStateNode *state.ReadChunkData) {
	reader := bufio.NewReader(msg.Response.Body)

	body, err := reader.ReadString('\n')
	if err == io.EOF {
		msg.Response.Body.Close()

		rcStateNode.Result = ReceiveStreamChunkMsg{}
		rcStateNode.Err = nil
		rcStateNode.Phase = async.DoneAsyncResultState
		return
	}
	if err != nil {
		rcStateNode.Result = ReceiveStreamChunkMsg{}
		rcStateNode.Err = err
		rcStateNode.Phase = async.DoneAsyncResultState
		return
	}

	body = strings.Trim(body, "\n")
	buf := []byte(body)

	if len(body) == 0 {
		return
	}

	err = msg.AgentChunk.ParseAgentChunk(&buf)
	if err != nil {
		rcStateNode.Result = ReceiveStreamChunkMsg{}
		rcStateNode.Err = err
		rcStateNode.Phase = async.DoneAsyncResultState
		return
	}

	rcStateNode.Result = msg
	rcStateNode.Err = nil
	rcStateNode.Phase = async.DoneAsyncResultState
}

func ProcessChunk(
	id string,
	processedChunkIds *[]string,
	render func() error,
) func(sink *[]string, chunk string) error {
	return func(sink *[]string, chunk string) error {
		if slices.Contains(*processedChunkIds, id) {
			return nil
		}

		(*sink)[len(*sink)-1] = (*sink)[len(*sink)-1] + chunk
		(*processedChunkIds) = append(*processedChunkIds, id)
		err := render()
		if err != nil {
			return err
		}
		return nil
	}
}
