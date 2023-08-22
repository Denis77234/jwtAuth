package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"medodsTest/internal/model"
)

type Mongo struct {
	coll *mongo.Collection
}

func New(uri string) (Mongo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return Mongo{}, fmt.Errorf("client initialization error:%w", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return Mongo{}, fmt.Errorf("client ping error:%w", err)
	}

	coll := client.Database("tokens").Collection("refresh")

	return Mongo{coll: coll}, nil
}

func (m *Mongo) Add(ctx context.Context, token model.Token) error {
	_, err := m.coll.InsertOne(ctx, token)
	if err != nil {
		return fmt.Errorf("insertion error:%w", err)
	}

	return nil
}

func (m *Mongo) Find(ctx context.Context, guid string, iat int64) (model.Token, error) {
	var refresh model.Token

	tokenEncode := m.coll.FindOne(ctx, bson.D{primitive.E{Key: "guid", Value: guid}, primitive.E{Key: "iat", Value: iat}})
	if tokenEncode.Err() != nil {
		return model.Token{}, fmt.Errorf("finding error:%w", tokenEncode.Err())
	}

	err := tokenEncode.Decode(&refresh)
	if err != nil {
		return model.Token{}, fmt.Errorf("decoding error:%w", tokenEncode.Err())
	}

	return refresh, nil
}

func (m *Mongo) Delete(ctx context.Context, guid string, iat int64) error {
	_, err := m.coll.DeleteOne(ctx, bson.D{primitive.E{Key: "guid", Value: guid}, primitive.E{Key: "iat", Value: iat}})
	if err != nil {
		return fmt.Errorf("deleting error:%w", err)
	}

	return nil
}

func (m *Mongo) Update(ctx context.Context, guid string, iat int64, upd model.Token) error {
	_, err := m.coll.UpdateOne(ctx, bson.D{primitive.E{Key: "guid", Value: guid}, primitive.E{Key: "iat", Value: iat}}, bson.M{"$set": upd})
	if err != nil {
		return fmt.Errorf("updating error:%w", err)
	}
	return nil
}

func (m *Mongo) Close(ctx context.Context) error {
	err := m.coll.Database().Client().Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}
