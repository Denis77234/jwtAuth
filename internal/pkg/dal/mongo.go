package dal

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"medosTest/internal/pkg/models"
	"time"
)

type Mongo struct {
	Client *mongo.Client
}

func New(uri string) (Mongo, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return Mongo{}, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return Mongo{}, err
	}

	return Mongo{Client: client}, nil
}

func (m *Mongo) AddRefresh(ctx context.Context, token models.Token) error {
	coll := m.Client.Database("tokens").Collection("refresh")

	_, err := coll.InsertOne(ctx, token)
	if err != nil {
		return err
	}
	return nil
}

func (m *Mongo) FindRefresh(ctx context.Context, guid string) (models.Token, error) {
	coll := m.Client.Database("tokens").Collection("refresh")

	var refresh models.Token

	tokenEncode := coll.FindOne(ctx, bson.D{{"guid", guid}})
	if tokenEncode.Err() != nil {
		return models.Token{}, nil
	}
	err := tokenEncode.Decode(&refresh)
	if err != nil {
		return models.Token{}, nil
	}
	return refresh, nil
}
