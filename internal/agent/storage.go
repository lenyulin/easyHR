package agent

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MessageRepository 定义消息存储的接口
// 提供保存消息、获取消息和删除消息的方法
// 是存储层的抽象，支持不同的存储实现

type MessageRepository interface {
	// SaveMessage 保存消息到存储
	// 接收上下文和消息，返回保存后的消息和错误
	SaveMessage(ctx context.Context, msg Message) (Message, error)

	// GetMessagesBySessionID 根据会话ID获取消息列表
	// 接收上下文和会话ID，返回消息列表和错误
	GetMessagesBySessionID(ctx context.Context, sessionID string) ([]Message, error)

	// DeleteMessage 删除消息
	// 接收上下文和消息ID，返回删除结果和错误
	DeleteMessage(ctx context.Context, id uint) (bool, error)

	// DeleteMessagesBySessionID 根据会话ID删除所有消息
	// 接收上下文和会话ID，返回删除结果和错误
	DeleteMessagesBySessionID(ctx context.Context, sessionID string) (bool, error)

	// GetMessageCount 获取消息总数
	// 接收上下文，返回消息总数和错误
	GetMessageCount(ctx context.Context) (int64, error)

	// Close 关闭存储连接
	// 接收上下文，返回关闭结果和错误
	Close(ctx context.Context) error
}

// StorageType 定义存储类型的枚举
// 支持MongoDB、MySQL等不同存储类型
// 用于创建不同的存储实例
type StorageType int

const (
	// StorageTypeMongoDB MongoDB存储类型
	StorageTypeMongoDB StorageType = iota

	// StorageTypeMySQL MySQL存储类型
	StorageTypeMySQL

	// StorageTypeMemory 内存存储类型（用于测试）
	StorageTypeMemory
)

// NewMessageRepository 创建一个新的MessageRepository实例
// 接收存储类型和配置，返回MessageRepository实例和错误
// 目前只实现了MongoDB存储，可扩展支持其他存储类型
func NewMessageRepository(storageType StorageType, config map[string]interface{}) (MessageRepository, error) {
	switch storageType {
	case StorageTypeMongoDB:
		// 创建MongoDB存储实例
		return NewMongoDBRepository(config)
	case StorageTypeMySQL:
		// 目前未实现MySQL存储，可扩展
		return nil, ErrStorageTypeNotSupported
	case StorageTypeMemory:
		// 目前未实现内存存储，可扩展
		return nil, ErrStorageTypeNotSupported
	default:
		return nil, ErrStorageTypeNotSupported
	}
}

// 自定义错误类型，用于存储操作
var (
	ErrStorageTypeNotSupported = Error("storage type not supported")
	ErrMessageNotFound         = Error("message not found")
	ErrFailedToSaveMessage     = Error("failed to save message")
	ErrFailedToGetMessages     = Error("failed to get messages")
	ErrFailedToDeleteMessage   = Error("failed to delete message")
)

// Error 自定义错误类型，实现error接口
// 用于存储操作中的错误处理

type Error string

// Error 实现error接口的Error方法
func (e Error) Error() string {
	return string(e)
}

// MongoDBRepository MongoDB消息存储的实现
// 包含MongoDB客户端、数据库和集合
// 实现了MessageRepository接口的所有方法

type MongoDBRepository struct {
	client     *mongo.Client     // MongoDB客户端
	db         *mongo.Database   // 数据库实例
	collection *mongo.Collection // 集合实例
}

// NewMongoDBRepository 创建一个新的MongoDBRepository实例
// 接收配置信息，返回MongoDBRepository实例和错误
// 配置信息包含MongoDB连接URL、数据库名称和集合名称
func NewMongoDBRepository(config map[string]interface{}) (*MongoDBRepository, error) {
	// 从配置中获取MongoDB连接URL
	connURL, ok := config["conn_url"].(string)
	if !ok {
		// 使用默认连接URL
		connURL = "mongodb://localhost:27017"
	}

	// 从配置中获取数据库名称
	dbName, ok := config["db_name"].(string)
	if !ok {
		// 使用默认数据库名称
		dbName = "ai_helper"
	}

	// 从配置中获取集合名称
	colName, ok := config["col_name"].(string)
	if !ok {
		// 使用默认集合名称
		colName = "messages"
	}

	// 设置MongoDB客户端选项
	opts := options.Client().ApplyURI(connURL)

	// 创建MongoDB客户端
	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// 检查连接是否成功
	if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// 获取数据库实例
	db := client.Database(dbName)

	// 获取集合实例
	collection := db.Collection(colName)

	// 返回MongoDBRepository实例
	return &MongoDBRepository{
		client:     client,
		db:         db,
		collection: collection,
	}, nil
}

// SaveMessage 保存消息到MongoDB
// 接收上下文和消息，返回保存后的消息和错误
func (r *MongoDBRepository) SaveMessage(ctx context.Context, msg Message) (Message, error) {
	// 检查消息的创建时间
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}

	// 保存消息到MongoDB
	result, err := r.collection.InsertOne(ctx, msg)
	if err != nil {
		return Message{}, fmt.Errorf("%w: %v", ErrFailedToSaveMessage, err)
	}

	// 将ObjectID转换为uint类型（仅用于演示，实际使用时可能需要调整）
	// 注意：MongoDB使用ObjectID作为默认ID，这里简化处理
	msg.ID = uint(result.InsertedID.(uint))

	return msg, nil
}

// GetMessagesBySessionID 根据会话ID获取消息列表
// 接收上下文和会话ID，返回消息列表和错误
func (r *MongoDBRepository) GetMessagesBySessionID(ctx context.Context, sessionID string) ([]Message, error) {
	// 构建查询条件
	filter := bson.M{"session_id": sessionID}

	// 设置排序选项
	sort := bson.D{{Key: "created_at", Value: 1}} // 按创建时间升序排列

	// 查询消息
	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(sort))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedToGetMessages, err)
	}
	defer cursor.Close(ctx)

	// 遍历结果
	var messages []Message
	for cursor.Next(ctx) {
		var msg Message
		if err := cursor.Decode(&msg); err != nil {
			return nil, fmt.Errorf("failed to decode message: %w", err)
		}
		messages = append(messages, msg)
	}

	// 检查遍历过程中是否发生错误
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedToGetMessages, err)
	}

	return messages, nil
}

// DeleteMessage 删除消息
// 接收上下文和消息ID，返回删除结果和错误
func (r *MongoDBRepository) DeleteMessage(ctx context.Context, id uint) (bool, error) {
	// 构建查询条件
	filter := bson.M{"_id": id}

	// 删除消息
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrFailedToDeleteMessage, err)
	}

	// 检查是否删除成功
	if result.DeletedCount == 0 {
		return false, ErrMessageNotFound
	}

	return true, nil
}

// DeleteMessagesBySessionID 根据会话ID删除所有消息
// 接收上下文和会话ID，返回删除结果和错误
func (r *MongoDBRepository) DeleteMessagesBySessionID(ctx context.Context, sessionID string) (bool, error) {
	// 构建查询条件
	filter := bson.M{"session_id": sessionID}

	// 删除消息
	result, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrFailedToDeleteMessage, err)
	}

	// 检查是否删除成功
	if result.DeletedCount == 0 {
		return false, ErrMessageNotFound
	}

	return true, nil
}

// GetMessageCount 获取消息总数
// 接收上下文，返回消息总数和错误
func (r *MongoDBRepository) GetMessageCount(ctx context.Context) (int64, error) {
	// 统计消息总数
	count, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("failed to get message count: %w", err)
	}

	return count, nil
}

// Close 关闭MongoDB连接
// 接收上下文，返回关闭结果和错误
func (r *MongoDBRepository) Close(ctx context.Context) error {
	// 关闭MongoDB客户端连接
	if err := r.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to close MongoDB connection: %w", err)
	}

	return nil
}
