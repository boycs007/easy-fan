# Easy-Fan

Easy-Fan 是一个基于 Go 语言实现的并行处理框架，采用 FAN IN/FAN OUT 模式，帮助开发者轻松实现并发任务处理。

## 特性

- 简单易用的 API 设计
- 支持泛型，类型安全
- 可配置的并发处理数量
- 可选的合并缓存大小
- 优雅的错误处理机制
- 支持上下文控制

## 安装

```bash
go get github.com/boycs007/easy-fan
```

## 快速开始

```go
package main

import (
    "context"
    "fmt"
    "github.com/boycs007/easy-fan"
)

func main() {
    ctx := context.Background()
    tasks := []int{1, 2, 3, 4, 5}
    
    processor := fan.GetProcessor(ctx, tasks, func(ctx context.Context, inChan <-chan int) <-chan fan.UnitRet[string] {
        out := make(chan fan.UnitRet[string])
        go func() {
            defer close(out)
            for num := range inChan {
                out <- fan.UnitRet[string]{
                    Item: fmt.Sprintf("Processed: %d", num),
                    Err:  nil,
                }
            }
        }()
        return out
    })

    results := processor.WithHandlerNum(3).Do()
    for _, result := range results {
        fmt.Println(result.Item)
    }
}
```

## 配置选项

### 并发处理数量

```go
processor.WithHandlerNum(5) // 设置5个并发处理器
```

### 合并缓存大小

```go
processor.WithMergeCache(100) // 设置合并通道的缓存大小为100
```

## 最佳实践

1. 合理设置并发数量，建议不要超过 CPU 核心数的 2 倍
2. 对于 I/O 密集型任务，可以适当增加并发数量
3. 使用 WithMergeCache 来优化处理速度差异较大的场景
4. 始终检查返回结果中的 Err 字段

---

# Easy-Fan

Easy-Fan is a parallel processing framework implemented in Go, utilizing the FAN IN/FAN OUT pattern to help developers easily implement concurrent task processing.

## Features

- Simple and intuitive API design
- Generic support with type safety
- Configurable concurrent processing count
- Optional merge cache size
- Elegant error handling mechanism
- Context support for cancellation and timeouts

## Installation

```bash
go get github.com/boycs007/easy-fan
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/boycs007/easy-fan"
)

func main() {
    ctx := context.Background()
    tasks := []int{1, 2, 3, 4, 5}
    
    processor := fan.GetProcessor(ctx, tasks, func(ctx context.Context, inChan <-chan int) <-chan fan.UnitRet[string] {
        out := make(chan fan.UnitRet[string])
        go func() {
            defer close(out)
            for num := range inChan {
                out <- fan.UnitRet[string]{
                    Item: fmt.Sprintf("Processed: %d", num),
                    Err:  nil,
                }
            }
        }()
        return out
    })

    results := processor.WithHandlerNum(3).Do()
    for _, result := range results {
        fmt.Println(result.Item)
    }
}
```

## Configuration Options

### Concurrent Processing Count

```go
processor.WithHandlerNum(5) // Set 5 concurrent processors
```

### Merge Cache Size

```go
processor.WithMergeCache(100) // Set merge channel buffer size to 100
```

## Best Practices

1. Set a reasonable number of concurrent processors, preferably not exceeding 2x the number of CPU cores
2. For I/O-bound tasks, consider increasing the concurrent count
3. Use WithMergeCache to optimize scenarios with varying processing speeds
4. Always check the Err field in the returned results

## License

MIT License 