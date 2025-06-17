package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

func GetConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err = json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	// Convert durations
	cfg.App.RTO = cfg.App.RTO * time.Millisecond
	cfg.App.WTO = cfg.App.WTO * time.Millisecond

	// Override with environment variables
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.App.Port = p
		}
	}

	// DB environment variables
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Repository.DBHost = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Repository.DBPort = p
		}
	}
	if user := os.Getenv("DB_USER"); user != "" {
		cfg.Repository.DBUsername = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		cfg.Repository.DBPassword = password
	}
	if name := os.Getenv("DB_NAME"); name != "" {
		cfg.Repository.DBName = name
	}

	// S3 environment variables
	if endpoint := os.Getenv("S3_ENDPOINT"); endpoint != "" {
		cfg.S3.Endpoint = endpoint
	}
	if accessKey := os.Getenv("S3_ACCESS_KEY"); accessKey != "" {
		cfg.S3.AccessKey = accessKey
	}
	if secretKey := os.Getenv("S3_SECRET_KEY"); secretKey != "" {
		cfg.S3.SecretKey = secretKey
	}
	if postBucket := os.Getenv("S3_POST_BUCKET"); postBucket != "" {
		cfg.S3.PostBucket = postBucket
	}
	if commentBucket := os.Getenv("S3_COMMENT_BUCKET"); commentBucket != "" {
		cfg.S3.CommentBucket = commentBucket
	}
	if useSSL := os.Getenv("S3_USE_SSL"); useSSL != "" {
		cfg.S3.UseSSL, _ = strconv.ParseBool(useSSL)
	}

	return &cfg, nil
}

type Config struct {
	App        App        `json:"app"`
	Repository Repository `json:"repository"`
	S3         S3         `json:"s3"`
}

type App struct {
	Port int           `json:"port"`
	RTO  time.Duration `json:"rto"`
	WTO  time.Duration `json:"wto"`
}

type Repository struct {
	DBHost      string `json:"db_host"`
	DBSrv       string `json:"db_srv"`
	DBPort      int    `json:"db_port"`
	DBUsername  string `json:"db_username"`
	DBPassword  string `json:"db_password"`
	DBName      string `json:"db_name"`
	DBSSLMode   string `json:"db_ssl_mode"`
	MaxConn     int    `json:"max_conn"`
	MaxIdleConn int    `json:"max_idle_conn"`
}

type S3 struct {
	Endpoint      string `json:"endpoint"`
	AccessKey     string `json:"access_key"`
	SecretKey     string `json:"secret_key"`
	PostBucket    string `json:"post_bucket"`
	CommentBucket string `json:"comment_bucket"`
	UseSSL        bool   `json:"use_ssl"`
}
