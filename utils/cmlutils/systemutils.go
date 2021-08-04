package cmlutils

import (
	"os"
)

func S3FullPath() string {
	return os.Getenv("BUCKET_PATH")
}

func S3ProxyPath() string {
	return os.Getenv("BUCKET_PROXY_PATH")
}

func S3BucketName() string {
	return os.Getenv("BUCKET_NAME")
}

func VOIPAPIPath() string {
	return os.Getenv("VOIP_API")
}

func CDRAPIPath() string {
	return os.Getenv("CDR_API")
}

func DefaultDatabase() string {
	return os.Getenv("DB_DEFAULT")
}

func EventsAPIPath() string {
	return os.Getenv("EVENT_API")
}

