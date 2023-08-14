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
	coll *mongo.Collection
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

	coll := client.Database("tokens").Collection("refresh")

	return Mongo{coll: coll}, nil
}

func (m *Mongo) Add(ctx context.Context, token models.Token) error {

	_, err := m.coll.InsertOne(ctx, token)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mongo) Find(ctx context.Context, guid string) (models.Token, error) {
	var refresh models.Token

	tokenEncode := m.coll.FindOne(ctx, bson.D{{"guid", guid}})
	if tokenEncode.Err() != nil {
		return models.Token{}, nil
	}

	err := tokenEncode.Decode(&refresh)
	if err != nil {
		return models.Token{}, nil
	}

	return refresh, nil
}

func (m *Mongo) Delete(ctx context.Context, guid string) error {

	_, err := m.coll.DeleteOne(ctx, bson.D{{"guid", guid}})
	if err != nil {
		return err
	}

	return nil
}

func (m *Mongo) Update(ctx context.Context, guid string, upd models.Token) error {
	_, err := m.coll.UpdateOne(ctx, bson.D{{"guid", guid}}, bson.M{"$set": upd})
	if err != nil {
		return err
	}
	return nil
}
