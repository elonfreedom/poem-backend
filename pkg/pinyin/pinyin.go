// Package pinyin 提供汉字转拼音功能
package pinyin

import (
	"strings"

	"github.com/mozillazg/go-pinyin"
)

var (
	// a 配置：带声调的拼音风格（jìng yè sī）
	a = pinyin.NewArgs()
)

func init() {
	a.Style = pinyin.Tone // 符号声调风格
}

// ToPinyin 将汉字转为拼音（带声调）
// 非汉字字符原样返回
func ToPinyin(s string) string {
	if s == "" {
		return ""
	}

	result := pinyin.Pinyin(s, a)
	var builder strings.Builder
	for i, p := range result {
		if i > 0 {
			builder.WriteByte(' ')
		}
		if len(p) > 0 {
			builder.WriteString(p[0])
		}
	}
	return builder.String()
}

// ToPinyinLines 将多行文本转为拼音（每行独立转换，保留换行）
func ToPinyinLines(s string) string {
	if s == "" {
		return ""
	}

	lines := strings.Split(s, "\n")
	pinyinLines := make([]string, 0, len(lines))
	for _, line := range lines {
		pinyinLines = append(pinyinLines, ToPinyin(line))
	}
	return strings.Join(pinyinLines, "\n")
}
