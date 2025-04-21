package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/daronenko/https-proxy/internal/app/config"
	"github.com/daronenko/https-proxy/internal/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrTransactionsNotFound = errors.New("transactions not found")
)

type Request struct {
	db   *mongo.Client
	conf *config.Config
}

func New(db *mongo.Client, conf *config.Config) *Request {
	return &Request{
		db:   db,
		conf: conf,
	}
}

func (repo *Request) CreateTransaction(ctx context.Context, transaction *model.Transaction) (*model.Transaction, error) {
	_, err := repo.getTransactionsCollection().InsertOne(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("creating http transaction error: %w", err)
	}

	return transaction, nil
}

func (repo *Request) GetTransactionByID(ctx context.Context, transactionID bson.ObjectID) (*model.Transaction, error) {
	filter := bson.M{"_id": transactionID}

	var transaction model.Transaction
	if err := repo.getTransactionsCollection().FindOne(ctx, filter).Decode(&transaction); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrTransactionsNotFound
		}
		return nil, fmt.Errorf("getting transaction by id error: %w", err)
	}

	return &transaction, nil
}

func (repo *Request) GetTransactionsList(ctx context.Context) ([]*model.Transaction, error) {
	cursor, err := repo.getTransactionsCollection().Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, fmt.Errorf("listing transactions error: %w", err)
	}
	defer cursor.Close(ctx)

	var results []*model.Transaction
	for cursor.Next(ctx) {
		var tx model.Transaction
		if err := cursor.Decode(&tx); err != nil {
			return nil, fmt.Errorf("decoding transaction error: %w", err)
		}
		results = append(results, &tx)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return results, nil
}

func (repo *Request) getTransactionsCollection() *mongo.Collection {
	return repo.db.Database(
		repo.conf.App.Mongo.Database,
	).Collection(
		repo.conf.App.Mongo.Collections.Transactions,
	)
}
