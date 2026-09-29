package response

import "github.com/gin-gonic/gin"

type DataEnvelope struct {
	Data any `json:"data"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

func WriteData(context *gin.Context, statusCode int, payload any) {
	context.JSON(statusCode, DataEnvelope{Data: payload})
}

func AbortWithError(context *gin.Context, statusCode int, errorCode, message string) {
	context.AbortWithStatusJSON(statusCode, ErrorEnvelope{Error: ErrorBody{Code: errorCode, Message: message}})
}
