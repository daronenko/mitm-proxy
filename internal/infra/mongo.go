package infra

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/daronenko/https-proxy/internal/app/config"
)

const (
	connectTimeout  = 30 * time.Second
	maxConnIdleTime = 3 * time.Minute
	minPoolSize     = 20
	maxPoolSize     = 300
)

func NewMongo(conf *config.Config) (*mongo.Client, error) {
	opts := options.Client().
		ApplyURI(conf.App.Mongo.URI).
		SetAuth(options.Credential{
			Username: conf.App.Mongo.Username,
			Password: conf.App.Mongo.Password,
		}).
		SetConnectTimeout(connectTimeout).
		SetMaxConnIdleTime(maxConnIdleTime).
		SetMinPoolSize(minPoolSize).
		SetMaxPoolSize(maxPoolSize)

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("connect to mongo error: %w", err)
	}

	if err := client.Ping(context.Background(), nil); err != nil {
		return nil, fmt.Errorf("ping mongo error: %w", err)
	}

	return client, nil
}
