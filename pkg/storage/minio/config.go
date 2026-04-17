package minio

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Secure    bool
	Bucket    string
	Region    string
	BaseURL   string
}
