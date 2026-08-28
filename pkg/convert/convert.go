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
