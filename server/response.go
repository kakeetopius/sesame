package server

// JSONResponse defines meta and interface struct
type JSONResponse struct {
	// Reserved field to add some meta information to the API response
	Meta any `json:"meta"`
	Data any `json:"data"`
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
