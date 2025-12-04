# ResumeScreen 模块设计文档
## 模块目标
该模块用于解析简历内容，并且提供搜索功能，使得用户可以快速找到符合要求的简历。
## 模块结构
resumescreen/
├── engine.go                 # [入口文件] 定义对外暴露的 Engine 接口和工厂方法
├── types.go                  # [领域模型] 定义 Candidate, ResumeData, SearchResult 等核心结构体
├── options.go                # [配置模式] 使用 Functional Options 模式配置权重、阈值
│
├── pipeline/                 # [录入子系统] 负责 ETL：OCR -> LLM 提取 -> 清洗
│   ├── processor.go          # 协调处理流程的主逻辑
│   ├── extractor.go          # LLM 结构化提取器 (JSON Parser)
│   ├── cleaner.go            # 去水分、归一化逻辑 (STAR 校验)
│   └── ocr_adapter.go        # 定义 OCR 接口 (具体实现由外部注入或内部封装)
│
├── search/                   # [检索子系统] 负责 RAG + SQL 混合检索
│   ├── service.go            # 搜索主逻辑
│   ├── query_builder.go      # NL2SQL：将自然语言转为 SQL + Vector Query
│   ├── ranker.go             # 混合排序算法 (Math/Score 计算)
│   └── filter.go             # 硬指标过滤器逻辑
│
├── evaluation/               # [评估子系统] 负责打分和价值判断
│   ├── rubric.go             # 评分标准定义 (数理化权重、学校排名计算)
│   └── verification.go       # 离群值检测与联网验证逻辑
│
└── storage/                  # [数据层抽象] 定义该模块需要的存储接口
    ├── repository.go         # 定义 ResumeRepository, VectorRepository 接口
    └── schema.go             # (可选) 定义该模块需要的 DB Table 结构
## 使用示例
```go
package main

import (
    "myproject/internal/llm"
    "myproject/internal/db"
    "myproject/pkg/resumescreen" // 引入子包
)

func main() {
    // 1. 准备依赖 (由主项目负责)
    myDB := db.NewPostgres(...)
    myLLM := llm.NewClient(...)
    
    // 2. 初始化筛选引擎 (注入依赖)
    screener := resumescreen.NewEngine(
        myLLM,
        myDB, 
        resumescreen.WithMathWeight(0.6), // 配置：侧重数学
    )

    // 3. 场景A：上传简历 (后台任务调用)
    screener.Ingest(ctx, "path/to/resume.pdf")

    // 4. 场景B：老师搜索 (HTTP Handler调用)
    candidates, _ := screener.Search(ctx, "我想要数理化好，且不偏科的学生")
}
```