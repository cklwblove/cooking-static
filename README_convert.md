# 图片转 WebP 工具

## 安装依赖

```bash
go mod download github.com/chai2010/webp
```

## 使用方法

### 基本用法（默认转换 ./static 目录）

```bash
go run convert_to_webp.go
```

### 参数选项

```bash
# 指定源目录
go run convert_to_webp.go -dir=./static/images

# 设置质量（0-100，默认85）
go run convert_to_webp.go -quality=90

# 保留原文件（默认会删除原文件）
go run convert_to_webp.go -keep

# 指定输出目录（保持原目录结构）
go run convert_to_webp.go -output=./output/webp

# 组合使用
go run convert_to_webp.go -dir=./static -output=./static_webp -quality=80 -keep
```

## 编译运行

```bash
# 编译
go build convert_to_webp.go

# 运行
./convert_to_webp -quality=85
```

## 功能特性

- 自动递归处理目录下所有 jpg、jpeg、png 文件（可通过修改常量扩展支持格式）
- 如果 webp 文件已存在会自动跳过
- 默认质量 85 可以在保持较高质量的同时大幅减小体积
- 默认会删除原文件，使用 `-keep` 参数可保留原文件
- 支持指定输出目录，自动保持原目录结构
- 显示跳过和失败文件的详细信息
- 转换结束后生成详细总结文件（`conversion_summary_*.txt`）
  - 统计数据（成功、跳过、失败数量）
  - 体积对比和压缩率
  - 每个文件的详细转换信息
  - 失败文件的具体原因

## 输出说明

### 控制台输出

转换过程中会实时显示：
- ✓ 成功转换的文件及压缩率
- ⊘ 跳过的文件
- ✗ 失败的文件及原因

转换结束后显示总结：
- 成功、跳过、失败数量统计
- 总体积对比和节省空间
- 跳过文件列表（最多显示 10 个）
- 失败文件列表及具体原因

### 总结文件

每次转换会生成一个带时间戳的总结文件，如 `conversion_summary_20251111_143025.txt`，包含完整的转换记录，方便后续查看和审计。

参考示例：`conversion_summary_example.txt`

## 扩展支持格式

如需支持更多图片格式，可修改代码顶部的常量：

```go
const (
	extJPG  = ".jpg"
	extJPEG = ".jpeg"
	extPNG  = ".png"
	extWEBP = ".webp"
)

var supportedExts = []string{extJPG, extJPEG, extPNG}
```

添加新格式后，需要在 `convertToWebP` 函数中增加相应的解码逻辑。

