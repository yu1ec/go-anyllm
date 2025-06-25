package response

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
)

const KEEP_ALIVE = `: keep-alive`

const KEEP_ALIVE_LEN = len(KEEP_ALIVE)

type StreamReader interface {
	Read() (*ChatCompletionsResponse, error)
	Next() bool
	Current() *ChatCompletionsResponse
	Error() error
}

type streamReader struct {
	respCh   chan *streamResponse
	current  *ChatCompletionsResponse
	err      error
	hasNext  bool
	finished bool
}

type streamResponse struct {
	chatResp *ChatCompletionsResponse
	error
}

func NewStreamReader(stream io.ReadCloser) StreamReader {
	iter := &streamReader{
		respCh:   make(chan *streamResponse),
		hasNext:  false,
		finished: false,
	}
	go iter.process(stream)
	return iter
}

func (m *streamReader) Read() (*ChatCompletionsResponse, error) {
	resp := <-m.respCh
	return resp.chatResp, resp.error
}

func (m *streamReader) Next() bool {
	if m.finished {
		return false
	}

	resp, ok := <-m.respCh
	if !ok {
		m.finished = true
		m.hasNext = false
		return false
	}

	m.current = resp.chatResp
	m.err = resp.error

	if resp.error != nil {
		if resp.error == io.EOF {
			m.finished = true
			m.hasNext = false
			m.err = nil
			return false
		} else {
			m.finished = true
			m.hasNext = false
			return false
		}
	}

	m.hasNext = true
	return true
}

func (m *streamReader) Current() *ChatCompletionsResponse {
	return m.current
}

func (m *streamReader) Error() error {
	return m.err
}

func (m *streamReader) process(stream io.ReadCloser) {
	defer stream.Close()
	defer close(m.respCh)

	reader := bufio.NewReader(stream)
	for {
		bytes, _, err := reader.ReadLine()
		if err != nil {
			m.respCh <- &streamResponse{nil, err}
			return
		}
		if len(bytes) <= 1 {
			continue
		}
		chatResp, err := processResponse(bytes)
		if err != nil {
			m.respCh <- &streamResponse{nil, err}
			return
		}
		m.respCh <- &streamResponse{chatResp, err}
	}
}

func processResponse(bytes []byte) (*ChatCompletionsResponse, error) {
	// handle keep-alive response
	if len(bytes) == KEEP_ALIVE_LEN {
		if string(bytes) == KEEP_ALIVE {
			err := errors.New("err: service unavailable")
			return nil, err
		}
	}

	// handle response end
	bytes = trimDataPrefix(bytes)
	if len(bytes) > 1 && bytes[0] == '[' {
		str := string(bytes)
		if str == "[DONE]" {
			return nil, io.EOF // io.EOF to indicate end
		}
	}

	// parse response
	chatResp := &ChatCompletionsResponse{}
	err := json.Unmarshal(bytes, chatResp)
	return chatResp, err
}

func trimDataPrefix(content []byte) []byte {
	trimIndex := 6
	if len(content) > trimIndex {
		return content[trimIndex:]
	}
	return content
}
