package video

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"videodetection/internal/config"
	"videodetection/internal/model"
)

type VideoInfo struct {
	Duration        float64 `json:"duration"`        // 视频时长(秒)
	FrameRate       float64 `json:"frameRate"`       // 帧率
	Width           int     `json:"width"`           // 宽度
	Height          int     `json:"height"`          // 高度
	Format          string  `json:"format"`          // 格式
	Bitrate         int64   `json:"bitrate"`         // 比特率(bps)
	Size            int64   `json:"size"`            // 文件大小(bytes)
	Codec           string  `json:"codec"`           // 视频编码
	AudioCodec      string  `json:"audioCodec"`      // 音频编码
	AudioChannels   int     `json:"audioChannels"`   // 音频通道数
	AudioSampleRate int     `json:"audioSampleRate"` // 音频采样率
}

type Processor struct {
	config *config.Config
}

func NewProcessor(cfg *config.Config) *Processor {
	return &Processor{
		config: cfg,
	}
}

// 获取视频信息
func (p *Processor) GetVideoInfo(videoPath string) (*VideoInfo, error) {
	log.Printf("开始获取视频信息: %s", videoPath)

	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		videoPath,
	)

	// 打印完整的 ffprobe 命令
	cmdStr := fmt.Sprintf("ffprobe %s", strings.Join(cmd.Args[1:], " "))
	log.Printf("执行 ffprobe 命令: %s", cmdStr)

	output, err := cmd.Output()
	if err != nil {
		log.Printf("ffprobe 命令执行失败: %v", err)
		return nil, fmt.Errorf("获取视频信息失败: %v", err)
	}

	var probeResult struct {
		Streams []struct {
			CodecType  string `json:"codec_type"`
			Width      int    `json:"width"`
			Height     int    `json:"height"`
			RFrameRate string `json:"r_frame_rate"`
			CodecName  string `json:"codec_name"`
			Channels   int    `json:"channels"`
			SampleRate string `json:"sample_rate"`
		} `json:"streams"`
		Format struct {
			Duration string `json:"duration"`
			Format   string `json:"format_name"`
			BitRate  string `json:"bit_rate"`
			Size     string `json:"size"`
		} `json:"format"`
	}

	if err := json.Unmarshal(output, &probeResult); err != nil {
		log.Printf("解析 ffprobe 输出失败: %v\n输出内容: %s", err, string(output))
		return nil, fmt.Errorf("解析视频信息失败: %v", err)
	}

	// 查找视频流和音频流
	var videoStream, audioStream *struct {
		CodecType  string `json:"codec_type"`
		Width      int    `json:"width"`
		Height     int    `json:"height"`
		RFrameRate string `json:"r_frame_rate"`
		CodecName  string `json:"codec_name"`
		Channels   int    `json:"channels"`
		SampleRate string `json:"sample_rate"`
	}

	for i := range probeResult.Streams {
		stream := &probeResult.Streams[i]
		if stream.CodecType == "video" && videoStream == nil {
			videoStream = stream
		} else if stream.CodecType == "audio" && audioStream == nil {
			audioStream = stream
		}
	}

	if videoStream == nil {
		log.Printf("未找到视频流，ffprobe 输出: %s", string(output))
		return nil, fmt.Errorf("未找到视频流")
	}

	// 解析帧率
	var frameRate float64
	if _, err := fmt.Sscanf(videoStream.RFrameRate, "%f", &frameRate); err != nil {
		frameRate = 0
	}

	// 解析时长
	var duration float64
	if _, err := fmt.Sscanf(probeResult.Format.Duration, "%f", &duration); err != nil {
		duration = 0
	}

	// 解析比特率和文件大小
	bitrate, _ := strconv.ParseInt(probeResult.Format.BitRate, 10, 64)
	size, _ := strconv.ParseInt(probeResult.Format.Size, 10, 64)

	info := &VideoInfo{
		Duration:  duration,
		FrameRate: frameRate,
		Width:     videoStream.Width,
		Height:    videoStream.Height,
		Format:    probeResult.Format.Format,
		Bitrate:   bitrate,
		Size:      size,
		Codec:     videoStream.CodecName,
	}

	// 如果有音频流，添加音频信息
	if audioStream != nil {
		info.AudioCodec = audioStream.CodecName
		info.AudioChannels = audioStream.Channels
		sampleRate, _ := strconv.Atoi(audioStream.SampleRate)
		info.AudioSampleRate = sampleRate
	}

	log.Printf("成功获取视频信息: %+v", info)
	return info, nil
}

// 抽帧方法
func (p *Processor) ExtractFrames(videoPath string, taskID string, options map[string]interface{}) ([]string, error) {
	log.Printf("开始抽帧任务 taskID=%s, videoPath=%s, options=%v", taskID, videoPath, options)

	outputDir := filepath.Join(p.config.Server.FrameDir, taskID)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Printf("创建输出目录失败 taskID=%s: %v", taskID, err)
		return nil, fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 构建 ffmpeg 命令
	args := []string{"-i", videoPath}

	// 获取图片格式
	imageFormat := "jpg"
	if format, ok := options["image_format"].(string); ok {
		log.Printf("收到图片格式参数: %s", format)
		imageFormat = format
	} else {
		log.Printf("未收到图片格式参数或格式不正确，使用默认格式: jpg")
	}

	var interval float64
	if intervalOpt, ok := options["interval"].(float64); ok {
		interval = intervalOpt
	} else if frameCount, ok := options["frame_count"].(int); ok {
		duration, err := p.GetVideoDuration(videoPath)
		if err != nil {
			return nil, err
		}
		interval = duration / float64(frameCount)
	}

	args = append(args, "-vf", fmt.Sprintf("fps=1/%f", interval))

	outputPattern := filepath.Join(outputDir, fmt.Sprintf("frame_%%d.%s", imageFormat))
	args = append(args, "-frame_pts", "1", outputPattern)

	// 根据格式添加编码参数
	switch imageFormat {
	case "png":
		args = append(args, "-compression_level", "6") // PNG压缩级别
	case "webp":
		args = append(args, "-quality", "90") // WebP质量
	case "jpg":
		args = append(args, "-q:v", "2") // JPEG质量
	}

	// 打印完整的 ffmpeg 命令
	cmdStr := fmt.Sprintf("ffmpeg %s", strings.Join(args, " "))
	log.Printf("执行 ffmpeg 命令 (使用 %s 格式): %s", imageFormat, cmdStr)

	cmd := exec.Command("ffmpeg", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("ffmpeg 命令执行失败: %v\n输出: %s", err, string(output))
		return nil, fmt.Errorf("抽帧失败: %v", err)
	}

	// 获取生成的帧文件列表
	frames, err := filepath.Glob(filepath.Join(outputDir, fmt.Sprintf("frame_*.%s", imageFormat)))
	if err != nil {
		log.Printf("获取帧文件列表失败: %v", err)
		return nil, fmt.Errorf("获取帧列表失败: %v", err)
	}

	// 计算每一帧的实际视频时间
	results := make([]*model.DetectionResult, len(frames))
	for i := range frames {
		videoTime := float64(i) * interval
		newPath := filepath.Join(outputDir, fmt.Sprintf("frame_%d_time_%.3f.%s", i, videoTime, imageFormat))
		if err := os.Rename(frames[i], newPath); err != nil {
			log.Printf("重命名帧文件失败: %v", err)
			continue
		}
		frames[i] = newPath
		results[i] = &model.DetectionResult{
			Type:        "frame",
			Timestamp:   float64(i),
			VideoTime:   videoTime,
			ImageFormat: imageFormat,
		}
	}

	log.Printf("抽帧完成 taskID=%s, 共生成 %d 帧", taskID, len(frames))
	return frames, nil
}

// 获取视频时长（秒）
func (p *Processor) GetVideoDuration(videoPath string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("获取视频时长失败: %v", err)
	}

	var duration float64
	if _, err := fmt.Sscanf(strings.TrimSpace(string(output)), "%f", &duration); err != nil {
		return 0, fmt.Errorf("解析视频时长失败: %v", err)
	}

	return duration, nil
}

// 清理临时文件
func (p *Processor) Cleanup(taskID string) error {
	dir := filepath.Join(p.config.Server.FrameDir, taskID)
	return os.RemoveAll(dir)
}
