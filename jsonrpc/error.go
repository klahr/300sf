package jsonrpc

type Error[T any] struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
}
