package server

// JSONResponse defines meta and interface struct
type JSONResponse struct {
	Status int `json:"status"`
	Data   any `json:"data"`
}

// JSONErrorResponse defines error struct
type JSONErrorResponse struct {
	Error *APIError `json:"error"`
}

// APIError defines api error struct
type APIError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}
