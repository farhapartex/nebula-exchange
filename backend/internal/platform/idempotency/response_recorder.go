package idempotency

import (
	"bytes"

	"github.com/gin-gonic/gin"
)

type responseRecorder struct {
	gin.ResponseWriter
	recordedBody bytes.Buffer
}

func (recorder *responseRecorder) Write(data []byte) (int, error) {
	recorder.recordedBody.Write(data)
	return recorder.ResponseWriter.Write(data)
}

func (recorder *responseRecorder) WriteString(text string) (int, error) {
	recorder.recordedBody.WriteString(text)
	return recorder.ResponseWriter.WriteString(text)
}
