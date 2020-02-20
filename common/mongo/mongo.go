package mongo

import (
	"context"
	"time"

	"github.com/golazycat/lazycron/common"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/golazycat/lazycron/common/baseconf"
	"go.mongodb.org/mongo-driver/mongo"
)

// mongodb连接器
type Connector struct {
	Client     *mongo.Client
	Database   *mongo.Database
	Collection *mongo.Collection
}

// 连接mongodb
func CreateConnect(conf *baseconf.MongoConf) (*Connector, error) {

	ctx, _ := context.WithTimeout(context.TODO(),
		time.Duration(conf.ConnectTimeout)*time.Second)

	client, err := mongo.Connect(ctx,
		options.Client().ApplyURI(conf.ConnectUrl))
	if err != nil {
		return nil, err
	}

	database := client.Database(common.MongodbDatabase)
	collection := database.Collection(common.MongodbCollection)

	return &Connector{
		Client:     client,
		Database:   database,
		Collection: collection,
	}, nil

}
