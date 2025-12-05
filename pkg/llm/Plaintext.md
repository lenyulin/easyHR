# LLM 模块设计文档
## 模块目标
该模块用于实现定义llm客户端，提供llm服务，供其他模块使用。
## 模块结构
llm/
├── llm.go                 # [入口文件] 定义对外暴露的 llm 接口和工厂方法
├── types.go                  # [领域模型] 定义核心结构体
│
├── manager/                 # [管理LLM] 负责新建、管理、销毁LLM客户端
│   ├── manager.go          # LLM 管理器
│   ├── openai_manager.go   # OpenAI LLM 管理器
│   ├── qwen_manager.go     # Qwen LLM 管理器
│
├── message/                   # [消息处理] 负责消息的添加、获取、持久化
│   ├── message.go            # 消息处理主逻辑
│
├── storage/                  # [数据层抽象] 定义该模块需要的存储接口
│   ├── repository.go         # 定义数据储存的接口
│   ├── storage.go            # 定义数据储存的实现，这里使用mongodb

## 工作机制
工作机制：
1. 实例创建：NewAIHelper构造函数初始化结构体，设置默认saveFunc为RabbitMQ发布。消息列表为空切片，SessionID从参数传入。
2. 消息添加：AddMessage方法添加新消息到历史，自动调用saveFunc持久化。若Save参数为false，仅内存存储。使用锁保护并发安全。
3. 响应生成：GenerateResponse/StreamResponse方法构建消息上下文，调用模型接口生成回复。用户消息先添加历史，AI回复后存储。流式模式通过回调实时输出。
4. 历史获取：GetMessages返回历史副本，避免外部修改。使用读锁确保线程安全。
5. 自定义存储：SetSaveFunc允许替换存储逻辑，如单元测试中的内存存储。
该架构通过组合模式和策略模式实现灵活性，支持多模型扩展和异步存储，适用于高并发的聊天应用。