package ipc

import (
	"encoding/json"
	"fmt"
	"net"
)

// Call sends a JSON-RPC 2.0 request to the daemon socket and returns the result.
func Call(socketPath, method string, params any) (json.RawMessage, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w (is mapitall running?)", socketPath, err)
	}
	defer conn.Close()

	var rawParams json.RawMessage
	if params != nil {
		rawParams, err = json.Marshal(params)
		if err != nil {
			return nil, err
		}
	}

	req := Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  rawParams,
		ID:      1,
	}

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return nil, err
	}

	var resp Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("rpc error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	raw, err := json.Marshal(resp.Result)
	if err != nil {
		return nil, err
	}
	return raw, nil
}
