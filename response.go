package main

import "net/http"

type Response struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func makeNotFoundError(
	message string, data interface{},
) (int, *Response) {

	return http.StatusNotFound, &Response{message, data}
}

func makeStatusOk(
	message string, data interface{},
) (int, *Response) {

	return http.StatusOK, &Response{message, data}
}

func makeInternalServerError(
	message string, data interface{},
) (int, *Response) {

	return http.StatusInternalServerError, &Response{message, data}
}
