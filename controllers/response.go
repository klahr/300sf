package controllers

type SingleItemDataResponse struct {
	Data interface{} `json:"data"`
}

type ListDataResponse struct {
	Data []interface{} `json:"data"`
}

type GeneralSuccessMessage struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type GeneralErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Code    int64  `json:"code"`
	} `json:"error"`
}

type ValidationErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Errors  []struct {
			Field   string `json:"field"`
			Message string `json:"message"`
		}
	} `json:"error"`
}
