package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ErrorHandler struct {
	Code        int    `json:"-"`
	Status      string `json:"status"`
	Message     string `json:"message,omitempty"`
	Description string `json:"description,omitempty"`
}

func NewErrorHandler(code int, err error, description string) *ErrorHandler {
	return &ErrorHandler{code, "failed", err.Error(), description}
}

func (he *ErrorHandler) Encode(rw http.ResponseWriter) {
	rw.WriteHeader(he.Code)
	encoder := json.NewEncoder(rw)
	encoder.Encode(he)
}

func SendSuccessful(rw http.ResponseWriter) {
	success := ErrorHandler{200, "successful", "", ""}
	success.Encode(rw)
}

func SendError(code int, err error, rw http.ResponseWriter) {
	fail := NewErrorHandler(code, err, "")
	fail.Encode(rw)
}

func SendJSONError(err error, reason string, rw http.ResponseWriter) {
	fail := ErrorHandler{
		http.StatusInternalServerError,
		"failed",
		err.Error(),
		"Error while " + reason + " JSON"}
	fail.Encode(rw)
}

var ErrUserNotFound = fmt.Errorf("User not found")
