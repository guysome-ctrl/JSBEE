package objectstore

import "strings"

const backupBucket = "backup"
const publicBucket = "public"
const privateBucket = "private"

const backupBaseName = "jsbee.sql.bak."

var buckets []string = []string{
	backupBucket,
	publicBucket,
	privateBucket,
}

// Public base URL for files in the public bucket, derived from
// the static config (path-style: `<scheme>://<url>/<publicBucket>`)
func publicBaseURLFromConfig(config s3StaticConfig) string {
	scheme := "https"
	if !config.Secure {
		scheme = "http"
	}

	return scheme + "://" + config.Url + "/" + publicBucket
}

// Content type for stored objects based on file extension
func contentTypeFor(filename string) string {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".pdf"):
		return "application/pdf"
	case strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".markdown"):
		return "text/markdown; charset=utf-8"
	case strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(lower, ".png"):
		return "image/png"
	case strings.HasSuffix(lower, ".webp"):
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
