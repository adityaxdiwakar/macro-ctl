package main

// CompInstruction: all instructions to be sent into websocket
type CompInstruction struct {
	Type string `json:"type"`
}

// OKResponse: all endpoints send 200
type OKResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
