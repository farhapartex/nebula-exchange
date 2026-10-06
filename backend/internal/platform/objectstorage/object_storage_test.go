package objectstorage

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/config"
)

func TestPresignedURLsPointAtThePublicEndpointWithoutCallingTheServer(t *testing.T) {
	storage, err := NewMinioStorage(config.StorageConfig{
		Endpoint:        "minio.internal:9000",
		PublicEndpoint:  "localhost:9000",
		AccessKey:       "street-born",
		SecretKey:       "street-born-secret",
		Bucket:          "street-born-assets",
		Region:          "us-east-1",
		PresignLifetime: 15 * time.Minute,
	})
	if err != nil {
		t.Fatalf("create storage: %v", err)
	}

	presignedURL, err := storage.PresignedDownloadURL(context.Background(), "story/1-1/05-fire.webp")
	if err != nil {
		t.Fatalf("presign: %v", err)
	}
	parsedURL, _ := url.Parse(presignedURL)
	if parsedURL.Host != "localhost:9000" || parsedURL.Path != "/street-born-assets/story/1-1/05-fire.webp" {
		t.Fatalf("unexpected url %s", presignedURL)
	}
	if parsedURL.Query().Get("X-Amz-Expires") != "900" || !strings.Contains(parsedURL.RawQuery, "X-Amz-Signature=") {
		t.Fatalf("expected a signed url valid for 900 seconds, got %s", presignedURL)
	}
}
