package jsonrpc

type Response[T any, U any] struct {
	JsonRpc string   `json:"jsonrpc"`
	Result  T        `json:"result,omitempty"`
	Error   Error[U] `json:"error,omitempty"`
	Id      int64    `json:"id"`
}
