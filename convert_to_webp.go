package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chai2010/webp"
)

const (
	extJPG  = ".jpg"
	extJPEG = ".jpeg"
	extPNG  = ".png"
	extWEBP = ".webp"
)

var supportedExts = []string{extJPG, extJPEG, extPNG}

type convertResult struct {
	path      string
	reason    string
	origSize  int64
	newSize   int64
	reduction float64
}

var (
	quality      = flag.Int("quality", 85, "webp 质量 (0-100)")
	keepOriginal = flag.Bool("keep", false, "是否保留原文件")
	targetDir    = flag.String("dir", "./static", "要处理的目录")
	outputDir    = flag.String("output", "", "输出目录（默认为源文件目录）")
)

func main() {
	flag.Parse()

	startTime := time.Now()

	fmt.Printf("开始转换图片...\n")
	fmt.Printf("源目录: %s\n", *targetDir)
	if *outputDir != "" {
		fmt.Printf("输出目录: %s\n", *outputDir)
	}
	fmt.Printf("质量: %d\n", *quality)
	fmt.Printf("保留原文件: %v\n\n", *keepOriginal)

	var successList []convertResult
	var skippedList []convertResult
	var failedList []convertResult

	var totalOrigSize, totalNewSize int64

	err := filepath.Walk(*targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !isSupportedExt(ext) {
			return nil
		}

		// 计算输出路径
		webpPath := getOutputPath(path, *targetDir, *outputDir)

		// 检查是否已有对应的 webp 文件
		if _, err := os.Stat(webpPath); err == nil {
			fmt.Printf("⊘ 已存在: %s\n", filepath.Base(webpPath))
			skippedList = append(skippedList, convertResult{
				path:   path,
				reason: "webp 文件已存在",
			})
			return nil
		}

		// 创建输出目录
		if err := os.MkdirAll(filepath.Dir(webpPath), 0755); err != nil {
			fmt.Printf("✗ 创建目录失败: %s (%v)\n", filepath.Dir(webpPath), err)
			failedList = append(failedList, convertResult{
				path:   path,
				reason: fmt.Sprintf("创建目录失败: %v", err),
			})
			return nil
		}

		if err := convertToWebP(path, webpPath, *quality); err != nil {
			fmt.Printf("✗ 失败: %s (%v)\n", filepath.Base(path), err)
			failedList = append(failedList, convertResult{
				path:   path,
				reason: err.Error(),
			})
			return nil
		}

		origSize := info.Size()
		newInfo, _ := os.Stat(webpPath)
		newSize := newInfo.Size()
		reduction := float64(origSize-newSize) / float64(origSize) * 100

		fmt.Printf("✓ %s → %s (%.1f%% 减少)\n",
			filepath.Base(path),
			filepath.Base(webpPath),
			reduction)

		successList = append(successList, convertResult{
			path:      path,
			origSize:  origSize,
			newSize:   newSize,
			reduction: reduction,
		})

		totalOrigSize += origSize
		totalNewSize += newSize

		if !*keepOriginal {
			os.Remove(path)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	duration := time.Since(startTime)

	// 打印总结
	printSummary(successList, skippedList, failedList, totalOrigSize, totalNewSize, duration)

	// 保存总结文件
	saveSummaryFile(successList, skippedList, failedList, totalOrigSize, totalNewSize, duration)
}

func isSupportedExt(ext string) bool {
	for _, supported := range supportedExts {
		if ext == supported {
			return true
		}
	}
	return false
}

func getOutputPath(inputPath, sourceDir, outputDir string) string {
	if outputDir == "" {
		// 输出到源文件目录
		ext := filepath.Ext(inputPath)
		return strings.TrimSuffix(inputPath, ext) + extWEBP
	}

	// 输出到指定目录，保持相对路径结构
	relPath, _ := filepath.Rel(sourceDir, inputPath)
	ext := filepath.Ext(relPath)
	webpRelPath := strings.TrimSuffix(relPath, ext) + extWEBP
	return filepath.Join(outputDir, webpRelPath)
}

func printSummary(successList, skippedList, failedList []convertResult, totalOrig, totalNew int64, duration time.Duration) {
	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf("转换总结\n")
	fmt.Printf(strings.Repeat("=", 60) + "\n\n")

	fmt.Printf("✓ 成功转换: %d 个文件\n", len(successList))
	fmt.Printf("⊘ 跳过: %d 个文件\n", len(skippedList))
	fmt.Printf("✗ 失败: %d 个文件\n\n", len(failedList))

	if len(successList) > 0 {
		totalReduction := float64(totalOrig-totalNew) / float64(totalOrig) * 100
		fmt.Printf("总体积: %.2f MB → %.2f MB\n",
			float64(totalOrig)/1024/1024,
			float64(totalNew)/1024/1024)
		fmt.Printf("节省空间: %.2f MB (%.1f%%)\n",
			float64(totalOrig-totalNew)/1024/1024,
			totalReduction)
	}

	fmt.Printf("耗时: %s\n\n", duration.Round(time.Millisecond))

	if len(skippedList) > 0 {
		fmt.Printf("跳过的文件:\n")
		for i, r := range skippedList {
			if i >= 10 {
				fmt.Printf("  ... 还有 %d 个文件\n", len(skippedList)-10)
				break
			}
			fmt.Printf("  - %s (%s)\n", r.path, r.reason)
		}
		fmt.Println()
	}

	if len(failedList) > 0 {
		fmt.Printf("失败的文件:\n")
		for _, r := range failedList {
			fmt.Printf("  - %s\n    原因: %s\n", r.path, r.reason)
		}
		fmt.Println()
	}
}

func saveSummaryFile(successList, skippedList, failedList []convertResult, totalOrig, totalNew int64, duration time.Duration) {
	filename := fmt.Sprintf("conversion_summary_%s.txt", time.Now().Format("20060102_150405"))

	f, err := os.Create(filename)
	if err != nil {
		fmt.Printf("⚠ 无法创建总结文件: %v\n", err)
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "图片转 WebP 格式转换总结\n")
	fmt.Fprintf(f, "生成时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(f, strings.Repeat("=", 80)+"\n\n")

	fmt.Fprintf(f, "统计数据:\n")
	fmt.Fprintf(f, "  成功转换: %d 个文件\n", len(successList))
	fmt.Fprintf(f, "  跳过文件: %d 个文件\n", len(skippedList))
	fmt.Fprintf(f, "  失败文件: %d 个文件\n", len(failedList))
	fmt.Fprintf(f, "  总耗时: %s\n\n", duration.Round(time.Millisecond))

	if len(successList) > 0 {
		totalReduction := float64(totalOrig-totalNew) / float64(totalOrig) * 100
		fmt.Fprintf(f, "体积对比:\n")
		fmt.Fprintf(f, "  转换前总体积: %.2f MB\n", float64(totalOrig)/1024/1024)
		fmt.Fprintf(f, "  转换后总体积: %.2f MB\n", float64(totalNew)/1024/1024)
		fmt.Fprintf(f, "  节省空间: %.2f MB (%.1f%%)\n\n",
			float64(totalOrig-totalNew)/1024/1024,
			totalReduction)
	}

	if len(successList) > 0 {
		fmt.Fprintf(f, strings.Repeat("-", 80)+"\n")
		fmt.Fprintf(f, "成功转换的文件 (%d):\n\n", len(successList))
		for _, r := range successList {
			fmt.Fprintf(f, "  %s\n", r.path)
			fmt.Fprintf(f, "    大小: %.2f KB → %.2f KB\n",
				float64(r.origSize)/1024,
				float64(r.newSize)/1024)
			fmt.Fprintf(f, "    压缩率: %.1f%%\n\n", r.reduction)
		}
	}

	if len(skippedList) > 0 {
		fmt.Fprintf(f, strings.Repeat("-", 80)+"\n")
		fmt.Fprintf(f, "跳过的文件 (%d):\n\n", len(skippedList))
		for _, r := range skippedList {
			fmt.Fprintf(f, "  %s\n", r.path)
			fmt.Fprintf(f, "    原因: %s\n\n", r.reason)
		}
	}

	if len(failedList) > 0 {
		fmt.Fprintf(f, strings.Repeat("-", 80)+"\n")
		fmt.Fprintf(f, "失败的文件 (%d):\n\n", len(failedList))
		for _, r := range failedList {
			fmt.Fprintf(f, "  %s\n", r.path)
			fmt.Fprintf(f, "    原因: %s\n\n", r.reason)
		}
	}

	fmt.Printf("✓ 总结文件已保存: %s\n", filename)
}

func convertToWebP(inputPath, outputPath string, quality int) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var img image.Image
	ext := strings.ToLower(filepath.Ext(inputPath))

	switch ext {
	case extJPG, extJPEG:
		img, err = jpeg.Decode(file)
	case extPNG:
		img, err = png.Decode(file)
	default:
		return fmt.Errorf("不支持的格式: %s", ext)
	}

	if err != nil {
		return err
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	return webp.Encode(out, img, &webp.Options{
		Lossless: false,
		Quality:  float32(quality),
	})
}
