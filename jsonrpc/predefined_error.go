package jsonrpc

var ParseError = Error[any]{
	Message: "Parse error",
	Code:    -32700,
}

var InvalidRequest = Error[any]{
	Message: "Invalid Request",
	Code:    -32600,
}

var MethodNotFound = Error[any]{
	Message: "Method not found",
	Code:    -32601,
}

var InvalidParams = Error[any]{
	Message: "Invalid params",
	Code:    -32602,
}

var InternalError = Error[any]{
	Message: "Internal error",
	Code:    -32603,
}
