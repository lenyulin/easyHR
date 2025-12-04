package storage

import (
	"context"
	"fmt"
)

// StorageType 定义存储类型
type StorageType string

const (
	// StorageTypeMongoDB MongoDB存储
	StorageTypeMongoDB StorageType = "mongodb"
	// StorageTypeLocal 本地文件存储
	StorageTypeLocal StorageType = "local"
	// StorageTypeOSS OSS存储
	StorageTypeOSS StorageType = "oss"
)

// Config 存储配置接口
type Config interface {
	GetType() StorageType
}

// Storage 定义通用存储接口
type Storage interface {
	// Insert 插入文档
	Insert(ctx context.Context, collection string, doc interface{}) error
	// Find 查询文档
	Find(ctx context.Context, collection string, filter interface{}) (interface{}, error)
	// Update 更新文档
	Update(ctx context.Context, collection string, filter interface{}, update interface{}) error
	// Delete 删除文档
	Delete(ctx context.Context, collection string, filter interface{}) error
	// Close 关闭存储连接
	Close(ctx context.Context) error
}

// NewStorage 创建存储实例
func NewStorage(config Config) (Storage, error) {
	switch config.GetType() {
	case StorageTypeMongoDB:
		mongoConfig, ok := config.(*MongoDBConfig)
		if !ok {
			return nil, fmt.Errorf("invalid MongoDB config type")
		}
		return newMongoDBStorage(mongoConfig)
	case StorageTypeLocal:
		// TODO: 实现本地存储
		return nil, fmt.Errorf("local storage not implemented yet")
	case StorageTypeOSS:
		// TODO: 实现OSS存储
		return nil, fmt.Errorf("oss storage not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.GetType())
	}
}

// InitStorage 初始化存储
func InitStorage(config Config) (Storage, error) {
	return NewStorage(config)
}
