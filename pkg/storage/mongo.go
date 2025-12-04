package storage

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBConfig MongoDB存储配置
type MongoDBConfig struct {
	URI           string
	Username      string
	Password      string
	DBName        string
	EnableMonitor bool
}

// GetType 获取存储类型
func (c *MongoDBConfig) GetType() StorageType {
	return StorageTypeMongoDB
}

type MongoStore struct {
	db *mongo.Database
}

// Insert 插入文档
func (s *MongoStore) Insert(ctx context.Context, collection string, doc interface{}) error {
	coll := s.db.Collection(collection)
	_, err := coll.InsertOne(ctx, doc)
	return err
}

// Find 查询文档
func (s *MongoStore) Find(ctx context.Context, collection string, filter interface{}) (interface{}, error) {
	coll := s.db.Collection(collection)
	var result interface{}
	err := coll.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Update 更新文档
func (s *MongoStore) Update(ctx context.Context, collection string, filter interface{}, update interface{}) error {
	coll := s.db.Collection(collection)
	_, err := coll.UpdateOne(ctx, filter, update)
	return err
}

// Delete 删除文档
func (s *MongoStore) Delete(ctx context.Context, collection string, filter interface{}) error {
	coll := s.db.Collection(collection)
	_, err := coll.DeleteOne(ctx, filter)
	return err
}

// Close 关闭存储连接
func (s *MongoStore) Close(ctx context.Context) error {
	return s.db.Client().Disconnect(ctx)
}

// newMongoDBStorage 创建MongoDB存储实例
func newMongoDBStorage(config *MongoDBConfig) (*MongoStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	opts := options.Client().
		ApplyURI(config.URI).
		SetAuth(options.Credential{
			Username: config.Username,
			Password: config.Password,
		})

	// 可选启用命令监控
	if config.EnableMonitor {
		monitor := &event.CommandMonitor{
			Started: func(ctx context.Context, startedEvent *event.CommandStartedEvent) {
				fmt.Println(startedEvent.Command)
			},
		}
		opts = opts.SetMonitor(monitor)
	}

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	// 如果未指定数据库名称，使用默认值
	dbName := config.DBName
	if dbName == "" {
		dbName = "easyhr"
	}

	return &MongoStore{
		db: client.Database(dbName),
	}, nil
}
