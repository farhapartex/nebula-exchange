package response

import (
	"github.com/gin-gonic/gin"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/pagination"
)

type DataEnvelope struct {
	Data any `json:"data"`
}

type ListEnvelope struct {
	Data       any                 `json:"data"`
	Pagination pagination.PageInfo `json:"pagination"`
}

type ErrorBody struct {
	Code    apierror.Code `json:"code"`
	Message string        `json:"message"`
	Details any           `json:"details,omitempty"`
}

type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

func WriteData(context *gin.Context, statusCode int, payload any) {
	context.JSON(statusCode, DataEnvelope{Data: payload})
}

func WriteList[Item any](context *gin.Context, statusCode int, page pagination.Page[Item]) {
	items := page.Items
	if items == nil {
		items = []Item{}
	}
	context.JSON(statusCode, ListEnvelope{Data: items, Pagination: page.Info})
}

func WriteError(context *gin.Context, err error) {
	apiError := apierror.From(err)
	if apiError.Code == apierror.CodeInternalError {
		_ = context.Error(err)
	}
	context.AbortWithStatusJSON(apiError.StatusCode, ErrorEnvelope{Error: ErrorBody{
		Code:    apiError.Code,
		Message: apiError.Message,
		Details: apiError.Details,
	}})
}
