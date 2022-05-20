package jsonrpc

type Request[T any] struct {
	JsonRpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []T    `json:"params,omitempty"`
	Id      int    `json:"id"`
}

