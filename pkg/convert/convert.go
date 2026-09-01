// Package convert 提供简繁体互转功能（基于 gocc，纯 Go 实现，无 C 依赖）
package convert

import (
	"fmt"
	"sync"

	"github.com/liuzl/gocc"
)

var (
	t2s     *gocc.OpenCC
	t2sOnce sync.Once
	t2sErr  error
	s2t     *gocc.OpenCC
	s2tOnce sync.Once
	s2tErr  error
)

// t2sConverter 返回单例的繁体转简体转换器
func t2sConverter() (*gocc.OpenCC, error) {
	t2sOnce.Do(func() {
		t2s, t2sErr = gocc.New("t2s")
	})
	return t2s, t2sErr
}

// s2tConverter 返回单例的简体转繁体转换器
func s2tConverter() (*gocc.OpenCC, error) {
	s2tOnce.Do(func() {
		s2t, s2tErr = gocc.New("s2t")
	})
	return s2t, s2tErr
}

// TraditionalToSimplified 将繁体中文转换为简体中文
func TraditionalToSimplified(in string) (string, error) {
	if in == "" {
		return "", nil
	}
	cc, err := t2sConverter()
	if err != nil {
		return "", fmt.Errorf("初始化繁转简转换器失败: %w", err)
	}
	out, err := cc.Convert(in)
	if err != nil {
		return "", fmt.Errorf("繁转简失败: %w", err)
	}
	return out, nil
}

// SimplifiedToTraditional 将简体中文转换为繁体中文
func SimplifiedToTraditional(in string) (string, error) {
	if in == "" {
		return "", nil
	}
	cc, err := s2tConverter()
	if err != nil {
		return "", fmt.Errorf("初始化简转繁转换器失败: %w", err)
	}
	out, err := cc.Convert(in)
	if err != nil {
		return "", fmt.Errorf("简转繁失败: %w", err)
	}
	return out, nil
}

// MustTraditionalToSimplified 繁转简（忽略错误，失败返回原文）
func MustTraditionalToSimplified(in string) string {
	out, err := TraditionalToSimplified(in)
	if err != nil {
		return in
	}
	return out
}

// MustSimplifiedToTraditional 简转繁（忽略错误，失败返回原文）
func MustSimplifiedToTraditional(in string) string {
	out, err := SimplifiedToTraditional(in)
	if err != nil {
		return in
	}
	return out
}

// CharsType 表示中文字符类型
type CharsType string

const (
	CharsTypeSimplified  CharsType = "simplified"   // 简体（含简体差异字，无繁体差异字）
	CharsTypeTraditional CharsType = "traditional"  // 繁体（含繁体差异字，无简体差异字）
	CharsTypeMixed       CharsType = "mixed"         // 混合（同时含简体和繁体差异字）
	CharsTypeNoDiff      CharsType = "no_diff"       // 无差异（有中文字符，但简繁体相同）
	CharsTypeUnknown     CharsType = "unknown"       // 未知（空字符串或无非中文字符）
)

// DetectCharsType 检测文本的中文字符类型
// 采用逐字分析，准确识别"部分字无差异、部分字有差异"的情况
//
// 返回值：
//   - "simplified"  : 含简体差异字（可能混有中性字），无繁体差异字
//   - "traditional" : 含繁体差异字（可能混有中性字），无简体差异字
//   - "mixed"       : 同时含简体差异字和繁体差异字
//   - "no_diff"     : 有中文字符，但所有字简繁体相同（无差异字）
//   - "unknown"     : 空字符串或无非中文字符
func DetectCharsType(in string) CharsType {
	if in == "" {
		return CharsTypeUnknown
	}

	hasDiffSimplified := false   // 是否存在简体差异字（简→繁会变）
	hasDiffTraditional := false  // 是否存在繁体差异字（繁→简会变）
	hasCJK := false              // 是否存在中文字符

	for _, r := range in {
		// 判断是否为 CJK 统一表意文字
		if !isCJK(r) {
			continue
		}
		hasCJK = true

		s := string(r)

		// 简→繁
		toTrad, err := SimplifiedToTraditional(s)
		if err == nil && toTrad != s {
			hasDiffSimplified = true
		}

		// 繁→简
		toSimp, err := TraditionalToSimplified(s)
		if err == nil && toSimp != s {
			hasDiffTraditional = true
		}

		// 提前退出：已确定混合
		if hasDiffSimplified && hasDiffTraditional {
			return CharsTypeMixed
		}
	}

	// 无中文字符
	if !hasCJK {
		return CharsTypeUnknown
	}

	switch {
	case hasDiffSimplified && hasDiffTraditional:
		return CharsTypeMixed
	case hasDiffSimplified:
		return CharsTypeSimplified
	case hasDiffTraditional:
		return CharsTypeTraditional
	default:
		// 有中文字符，但所有字简繁体相同
		return CharsTypeNoDiff
	}
}

// isCJK 判断字符是否为 CJK 统一表意文字
func isCJK(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified Ideographs
		(r >= 0x3400 && r <= 0x4DBF) || // CJK Extension A
		(r >= 0x20000 && r <= 0x2A6DF) // CJK Extension B
}

// ConvertChars 将文本转换为目标字符类型
// target: "simplified" 或 "traditional"
func ConvertChars(in string, target string) (string, error) {
	if in == "" {
		return "", nil
	}
	switch target {
	case "simplified":
		return TraditionalToSimplified(in)
	case "traditional":
		return SimplifiedToTraditional(in)
	default:
		return "", fmt.Errorf("不支持的目标类型: %s", target)
	}
}
