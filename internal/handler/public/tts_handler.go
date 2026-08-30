package public

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-fuego/fuego"

	"poem-backend/pkg/response"
)

// flusher 接口：用于强制刷新 HTTP 响应缓冲区，实现流式传输
type flusher interface {
	http.ResponseWriter
	Flush()
}

// ensureFlusher 将 ResponseWriter 包装为支持 Flush 的类型
type ensureFlusher struct {
	http.ResponseWriter
}

func (e *ensureFlusher) Flush() {
	if f, ok := e.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

const maxTextLength = 5000 // 最大字符数限制

// TTSHandler TTS 朗读处理器
type TTSHandler struct {
	piperURL      string // Piper 服务地址，如 http://localhost:5000
	cacheDir      string // 本地缓存目录
	rateLimiter   *rateLimiter
	piperSem      chan struct{} // Piper 并发限制
}

// NewTTSHandler 创建 TTS 处理器
func NewTTSHandler() *TTSHandler {
	piperHost := getEnv("PIPER_HOST", "localhost")
	piperPort := getEnv("PIPER_PORT", "5000")
	cacheDir := getEnv("TTS_CACHE_DIR", "/tmp/tts-cache")

	// 确保缓存目录存在
	_ = os.MkdirAll(cacheDir, 0755)

	return &TTSHandler{
		piperURL:    fmt.Sprintf("http://%s:%s", piperHost, piperPort),
		cacheDir:    cacheDir,
		rateLimiter: newRateLimiter(20, time.Minute), // 20 req/min
		piperSem:    make(chan struct{}, 3),           // Piper 最大并发 3
	}
}

// TTSSentenceTimestamp 句子时间戳
type TTSSentenceTimestamp struct {
	Text  string  `json:"text" description:"句子文本"`
	Start float64 `json:"start" description:"开始时间（秒）"`
	End   float64 `json:"end" description:"结束时间（秒）"`
}

// TTimestampsResponse 时间戳响应
type TTimestampsResponse struct {
	Sentences []TTSSentenceTimestamp `json:"sentences" description:"句子时间戳列表"`
	Total     float64                `json:"total" description:"总时长（秒）"`
}

// Timestamps 获取句子时间戳接口（保留向后兼容，新客户端建议直接使用 /tts 的 X-Timestamps 头）
// GET /api/public/tts/timestamps?text=床前明月光，疑是地上霜。举头望明月，低头思故乡。
func (h *TTSHandler) Timestamps(c fuego.ContextNoBody) (*response.APIResponse[TTimestampsResponse], error) {
	text := c.QueryParam("text")
	if text == "" {
		return nil, fuego.BadRequestError{Title: "missing text", Detail: "text 参数不能为空"}
	}
	if len(text) > maxTextLength {
		return nil, fuego.BadRequestError{Title: "text too long", Detail: fmt.Sprintf("text 长度不能超过 %d 字符", maxTextLength)}
	}

	// 解析语速
	lengthScale := 1.0
	if speed := c.QueryParam("speed"); speed != "" {
		if s, err := strconv.ParseFloat(speed, 64); err == nil && s > 0 {
			lengthScale = s
		}
	}

	// 按句分割
	sentences := splitSentences(text)
	if len(sentences) == 0 {
		return nil, fuego.BadRequestError{Title: "no sentences", Detail: "未检测到有效句子"}
	}

	// 计算时间戳（仅从缓存获取，不调用 Piper，确保快速响应）
	timestamps := h.computeTimestampsFromCache(text, lengthScale)
	if timestamps == nil {
		return nil, fuego.InternalServerError{Title: "tts unavailable", Detail: "语音合成服务暂不可用"}
	}

	return response.OK(TTimestampsResponse{
		Sentences: timestamps,
		Total:     round2(calcTotal(timestamps)),
	}), nil
}

// calcTotal 计算时间戳总时长
func calcTotal(timestamps []TTSSentenceTimestamp) float64 {
	if len(timestamps) == 0 {
		return 0
	}
	return timestamps[len(timestamps)-1].End
}

// Synthesize 语音合成接口（流式 WebM/Opus + 时间戳）
// GET /api/public/tts?text=床前明月光&speed=1.0
// 返回 Transfer-Encoding: chunked 流式音频数据
// 响应头 X-Timestamps 携带 base64 编码的句子时间戳 JSON
func (h *TTSHandler) Synthesize(c fuego.ContextNoBody) (*response.APIResponse[any], error) {
	text := c.QueryParam("text")
	if text == "" {
		return nil, fuego.BadRequestError{Title: "missing text", Detail: "text 参数不能为空"}
	}
	if len(text) > maxTextLength {
		return nil, fuego.BadRequestError{Title: "text too long", Detail: fmt.Sprintf("text 长度不能超过 %d 字符", maxTextLength)}
	}

	// 限流检查（基于 IP，去掉端口）
	clientIP := c.Request().RemoteAddr
	if idx := strings.LastIndex(clientIP, ":"); idx >= 0 {
		clientIP = clientIP[:idx]
	}
	if !h.rateLimiter.allow(clientIP) {
		c.Response().Header().Set("Content-Type", "application/json")
		c.Response().WriteHeader(http.StatusTooManyRequests)
		_, _ = c.Response().Write([]byte(`{"code":429,"message":"rate limited","error":"请求过于频繁，请稍后再试（限制：20次/分钟）"}`))
		return nil, nil
	}

	// 解析语速
	lengthScale := 1.0
	if speed := c.QueryParam("speed"); speed != "" {
		if s, err := strconv.ParseFloat(speed, 64); err == nil && s > 0 {
			lengthScale = s
		}
	}

	// 计算缓存 key（fMP4 格式）
	cacheKey := h.cacheKey(text, lengthScale)
	cachePath := filepath.Join(h.cacheDir, cacheKey+".m4a")

	// 检查 fMP4 缓存（24h TTL）— 此时已知时间戳，直接返回
	if data, ok := h.getCache(cachePath); ok {
		// 从缓存计算时间戳（仅读 WAV 缓存，不调用 Piper）
		timestamps := h.computeTimestampsFromCache(text, lengthScale)
		tsHeader := encodeTimestampsHeader(timestamps)
		c.Response().Header().Set("Content-Type", "audio/mp4; codecs=opus")
		c.Response().Header().Set("X-Cache", "HIT")
		c.Response().Header().Set("X-Timestamps", tsHeader)
		c.Response().Header().Set("Cache-Control", "public, max-age=86400")
		_, _ = c.Response().Write(data)
		return nil, nil
	}

	// 流式合成：逐句 Piper WAV → ffmpeg fMP4/Opus → chunked HTTP response
	// 时间戳从缓存获取（仅读已缓存的 WAV），未缓存的句子在流式合成时计算
	timestamps := h.computeTimestampsFromCache(text, lengthScale)
	tsHeader := encodeTimestampsHeader(timestamps)

	return h.streamAudio(c, text, lengthScale, cachePath, tsHeader)
}

// computeTimestampsFromCache 仅从缓存计算句子时间戳（不调用 Piper）
// 未缓存的句子使用估算时长，确保流式传输可以立即开始
func (h *TTSHandler) computeTimestampsFromCache(text string, lengthScale float64) []TTSSentenceTimestamp {
	sentences := splitSentences(text)
	if len(sentences) == 0 {
		return nil
	}

	var timestamps []TTSSentenceTimestamp
	var currentTime float64
	sentenceCacheDir := h.cacheDir + "/sentences"

	for _, sentence := range sentences {
		sentenceStart := currentTime
		duration := 0.0

		// 尝试从缓存读取
		sentenceKey := h.cacheKey(sentence, lengthScale)
		sentenceCachePath := filepath.Join(sentenceCacheDir, sentenceKey+".wav")
		if cached, ok := h.getCache(sentenceCachePath); ok {
			duration = h.getWAVDuration(cached)
			if duration <= 0 {
				duration = float64(len(cached)-44) / 44100.0 // Piper: 22050Hz * 2 bytes
				if duration < 0 {
					duration = 0
				}
			}
		} else {
			// 未缓存：按字数估算（中文约 0.2s/字）
			duration = float64(len([]rune(sentence))) * 0.2
		}

		timestamps = append(timestamps, TTSSentenceTimestamp{
			Text:  sentence,
			Start: round2(sentenceStart),
			End:   round2(sentenceStart + duration),
		})

		// 句间停顿 200ms
		currentTime = sentenceStart + duration + 0.2
	}

	return timestamps
}

// encodeTimestampsHeader 将时间戳编码为 base64 JSON 字符串，用于 HTTP 响应头
func encodeTimestampsHeader(timestamps []TTSSentenceTimestamp) string {
	if len(timestamps) == 0 {
		return ""
	}
	data, err := json.Marshal(timestamps)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(data)
}

// decodeTimestampsHeader 从 base64 编码的 HTTP 响应头解析时间戳（前端参考用）
func decodeTimestampsHeader(header string) []TTSSentenceTimestamp {
	if header == "" {
		return nil
	}
	data, err := base64.StdEncoding.DecodeString(header)
	if err != nil {
		return nil
	}
	var timestamps []TTSSentenceTimestamp
	if err := json.Unmarshal(data, &timestamps); err != nil {
		return nil
	}
	return timestamps
}

// streamAudio 真正的流式音频合成：
// 逐句调用 Piper /synthesize → 提取 raw PCM → ffmpeg 实时转码 → HTTP chunked response
// 首字节时间 = 第一句合成时间（~200ms），与全文长度无关
func (h *TTSHandler) streamAudio(c fuego.ContextNoBody, text string, lengthScale float64, cachePath, tsHeader string) (*response.APIResponse[any], error) {
	ctx := c.Context()

	// 获取 Piper 并发信号量
	select {
	case h.piperSem <- struct{}{}:
		defer func() { <-h.piperSem }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// 按句分割
	sentences := splitSentences(text)
	if len(sentences) == 0 {
		return nil, fuego.BadRequestError{Title: "no sentences", Detail: "未检测到有效句子"}
	}

	// 设置响应头（立即发送，不等待合成）
	// 使用 fMP4 (fragmented MP4) + Opus：
	// - WebM 管道输出时 EBML Segment Size 无法回填，浏览器 demuxer 解析失败
	// - Ogg/Opus 在 iOS/Android 上不支持
	// - fMP4 + Opus 所有浏览器原生支持，且支持流式管道输出
	c.Response().Header().Set("Content-Type", "audio/mp4; codecs=opus")
	c.Response().Header().Set("X-Cache", "MISS")
	c.Response().Header().Set("X-Timestamps", tsHeader)
	c.Response().Header().Set("Cache-Control", "public, max-age=86400")
	c.Response().WriteHeader(http.StatusOK)

	fw := &ensureFlusher{ResponseWriter: c.Response()}

	// 启动 ffmpeg：raw PCM pipe in → fMP4/Opus pipe out
	// Piper 输出 22050Hz 16bit 单声道 PCM
	ffmpegCmd := exec.CommandContext(ctx, "ffmpeg",
		"-hide_banner", "-loglevel", "error",
		"-f", "s16le", // raw PCM 16-bit little-endian
		"-ar", "22050", // Piper 原生采样率
		"-ac", "1", // 单声道
		"-i", "pipe:0", // stdin
		"-c:a", "libopus",
		"-b:a", "16k", // 16kbps
		"-ar", "48000", // Opus 原生采样率 48kHz，浏览器自行重采样
		"-ac", "1",
		"-application", "voip",
		"-f", "mp4", // MP4 容器
		"-movflags", "empty_moov+frag_keyframe+default_base_moof", // fragmented MP4 for streaming
		"pipe:1", // stdout
	)

	ffmpegIn, err := ffmpegCmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg stdin pipe: %w", err)
	}

	ffmpegOut, err := ffmpegCmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg stdout pipe: %w", err)
	}

	if err := ffmpegCmd.Start(); err != nil {
		return nil, fmt.Errorf("ffmpeg start: %w", err)
	}

	// 创建临时缓存文件，边流式传输边写入缓存
	tmpCache, cacheWriteErr := os.CreateTemp(h.cacheDir, "tts-stream-*.m4a.tmp")
	var cacheFile *os.File
	if cacheWriteErr == nil {
		cacheFile = tmpCache
		defer func() {
			cacheFile.Close()
			// 原子移动到最终位置（仅当文件有内容时）
			if info, statErr := os.Stat(cacheFile.Name()); statErr == nil && info.Size() > 0 {
				_ = os.Rename(cacheFile.Name(), cachePath)
			} else {
				os.Remove(cacheFile.Name())
			}
		}()
	}

	// 同时写入 HTTP response 和缓存文件
	writers := []io.Writer{fw}
	if cacheFile != nil {
		writers = append(writers, cacheFile)
	}
	mw := io.MultiWriter(writers...)

	// 后台：逐句合成 → 提取 PCM → 写入 ffmpeg stdin
	sentenceCacheDir := h.cacheDir + "/sentences"
	_ = os.MkdirAll(sentenceCacheDir, 0755)

	go func() {
		for i, sentence := range sentences {
			// 检查单句缓存
			sentenceKey := h.cacheKey(sentence, lengthScale)
			sentenceCachePath := filepath.Join(sentenceCacheDir, sentenceKey+".wav")

			var wavData []byte
			if cached, ok := h.getCache(sentenceCachePath); ok {
				wavData = cached
			} else {
				var callErr error
				wavData, callErr = h.callPiper(ctx, sentence, lengthScale)
				if callErr != nil {
					break // Piper 不可用，停止
				}
				h.setCache(sentenceCachePath, wavData)
			}

			// 提取 raw PCM（去掉 WAV 头）
			pcmData := stripWavHeader(wavData)

			// 写入 ffmpeg stdin
			if _, writeErr := ffmpegIn.Write(pcmData); writeErr != nil {
				break
			}

			// 句间停顿 200ms（除了最后一句）
			if i < len(sentences)-1 {
				silenceSamples := int(22050 * 0.2) // 22050Hz * 0.2s
				silence := make([]byte, silenceSamples*2) // 16bit = 2 bytes/sample
				if _, writeErr := ffmpegIn.Write(silence); writeErr != nil {
					break
				}
			}
		}
		ffmpegIn.Close() // 通知 ffmpeg 输入结束
	}()

	// 主线程：ffmpeg stdout → HTTP chunked response + cache（边转码边发送）
	buf := make([]byte, 16384) // 16KB 缓冲区
	for {
		n, readErr := ffmpegOut.Read(buf)
		if n > 0 {
			_, _ = mw.Write(buf[:n])
			fw.Flush() // 强制刷新，立即发送给客户端
		}
		if readErr != nil {
			break
		}
	}

	// 等待 ffmpeg 完成
	_ = ffmpegCmd.Wait()

	return nil, nil
}

// stripWavHeader 从 WAV 数据中提取 raw PCM（去掉 RIFF/WAVE 头）
func stripWavHeader(wavData []byte) []byte {
	if len(wavData) < 44 {
		return wavData
	}
	// 查找 "data" chunk（WAV 格式中音频数据从 data chunk 开始）
	for i := 12; i < len(wavData)-8; i++ {
		if wavData[i] == 'd' && wavData[i+1] == 'a' && wavData[i+2] == 't' && wavData[i+3] == 'a' {
			dataSize := int(wavData[i+4]) | int(wavData[i+5])<<8 | int(wavData[i+6])<<16 | int(wavData[i+7])<<24
			end := i + 8 + dataSize
			if end > len(wavData) {
				end = len(wavData)
			}
			return wavData[i+8 : end]
		}
	}
	// Fallback: 标准 44 字节头
	return wavData[44:]
}

// getWAVDuration 获取 WAV 数据时长（秒），使用 ffprobe
func (h *TTSHandler) getWAVDuration(wavData []byte) float64 {
	tmpFile, err := os.CreateTemp("", "tts-dur-*.wav")
	if err != nil {
		return 0
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Write(wavData)
	tmpFile.Close()

	cmd := exec.Command("ffprobe", "-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		tmpFile.Name(),
	)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	var duration float64
	fmt.Sscanf(string(output), "%f", &duration)
	return duration
}

// callPiper 调用 Piper HTTP 服务合成语音（带并发限制）
func (h *TTSHandler) callPiper(ctx context.Context, text string, lengthScale float64) ([]byte, error) {
	// 获取 Piper 并发信号量
	select {
	case h.piperSem <- struct{}{}:
		defer func() { <-h.piperSem }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	reqBody := fmt.Sprintf(`{"text": %q, "length_scale": %g}`, text, lengthScale)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.piperURL+"/synthesize",
		strings.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("piper returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// cacheKey 生成缓存 key
func (h *TTSHandler) cacheKey(text string, lengthScale float64) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s|%.2f", text, lengthScale)))
	return hex.EncodeToString(hash[:])
}

// getCache 获取缓存（24h TTL）
func (h *TTSHandler) getCache(path string) ([]byte, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, false
	}
	// 24h TTL
	if time.Since(info.ModTime()) > 24*time.Hour {
		_ = os.Remove(path)
		return nil, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	return data, true
}

// setCache 写入缓存
func (h *TTSHandler) setCache(path string, data []byte) {
	_ = os.WriteFile(path, data, 0644)
}

// splitSentences 按标点分割句子（用于高亮）
// 古诗朗读场景：，。！？； 都是自然停顿，都作为断句点
func splitSentences(text string) []string {
	var sentences []string

	// 先按行分割（换行符是天然的分隔）
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 按标点分割：，。！？；
		current := ""
		for _, r := range line {
			current += string(r)
			if r == '，' || r == '。' || r == '！' || r == '？' || r == '；' {
				s := strings.TrimSpace(current)
				if s != "" {
					sentences = append(sentences, s)
				}
				current = ""
			}
		}

		// 行末无标点的剩余部分
		if remainder := strings.TrimSpace(current); remainder != "" {
			sentences = append(sentences, remainder)
		}
	}

	return sentences
}


// round2 保留两位小数
func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// ========== 限流器 ==========

type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int
	window   time.Duration
}

type visitor struct {
	count    int
	lastSeen time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}
	// 启动清理 goroutine
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[key]
	if !exists || now.Sub(v.lastSeen) > rl.window {
		rl.visitors[key] = &visitor{count: 1, lastSeen: now}
		return true
	}

	if v.count >= rl.limit {
		return false
	}
	v.count++
	v.lastSeen = now
	return true
}

func (rl *rateLimiter) cleanup() {
	for {
		time.Sleep(rl.window)
		rl.mu.Lock()
		now := time.Now()
		for k, v := range rl.visitors {
			if now.Sub(v.lastSeen) > rl.window {
				delete(rl.visitors, k)
			}
		}
		rl.mu.Unlock()
	}
}
