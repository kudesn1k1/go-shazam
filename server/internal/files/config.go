package files

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Endpoint        string
	Region          string
	AccessKey       string
	SecretKey       string
	Bucket          string
	UsePathStyle    bool
	MaxUploadBytes  int64
	CleanupOutdate  time.Duration
	CleanupSchedule string
}

func LoadConfig() *Config {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()
	viper.AutomaticEnv()

	viper.SetDefault("S3_ENDPOINT", "http://minio:9000")
	viper.SetDefault("S3_REGION", "us-east-1")
	viper.SetDefault("S3_BUCKET", "go-shazam")
	viper.SetDefault("S3_USE_PATH_STYLE", true)
	viper.SetDefault("FILE_MAX_UPLOAD_BYTES", int64(2*1024*1024))
	viper.SetDefault("FILE_CLEANUP_OUTDATE", "720h")
	viper.SetDefault("FILE_CLEANUP_SCHEDULE", "@weekly")

	outdate, err := time.ParseDuration(viper.GetString("FILE_CLEANUP_OUTDATE"))
	if err != nil {
		outdate = 720 * time.Hour
	}

	return &Config{
		Endpoint:        viper.GetString("S3_ENDPOINT"),
		Region:          viper.GetString("S3_REGION"),
		AccessKey:       viper.GetString("S3_ACCESS_KEY"),
		SecretKey:       viper.GetString("S3_SECRET_KEY"),
		Bucket:          viper.GetString("S3_BUCKET"),
		UsePathStyle:    viper.GetBool("S3_USE_PATH_STYLE"),
		MaxUploadBytes:  viper.GetInt64("FILE_MAX_UPLOAD_BYTES"),
		CleanupOutdate:  outdate,
		CleanupSchedule: viper.GetString("FILE_CLEANUP_SCHEDULE"),
	}
}
