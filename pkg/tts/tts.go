package tts

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// 传入30s闹钟,文案，和保存路径，调用edge-tts生成mp3
// generateaudio = tts = text to speech
func GenerateAudio(ctx context.Context, text string, outputPath string) error {
	//自动创建保存音频的父级目录,权限为0755
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}

	//定义临时纯人声文件路径
	rawVoice := outputPath + ".raw.mp3"

	//结束函数退出时删除临时文件
	defer os.Remove(rawVoice)

	//根据音频路径推导出字幕路径
	subPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".vtt"
	//调用edge-tts生成纯人声,加上 --rate=15%压慢语速
	//--writer-subtitles自动导出字幕文件
	ttsCmd := exec.CommandContext(ctx, "edge-tts", "--text", text, "--write-media", rawVoice, "--write-subtitles", subPath, "--voice", "zh-CN-YunxiNeural", "--rate=-15%")
	if err := ttsCmd.Run(); err != nil {
		return err
	}

	//组装滤镜:把bgm音量压低到15%,混音并以人声长度为准 自动截断
	filter := "[1:a]volume=0.15[bgm];[0:a][bgm]amix=inputs=2:duration=first:dropout_transition=2"

	//启动ffmpeg子进程:-y覆盖同名文件，-i输出人声与bgm，执行滤镜并用libmp3lame编码器混流导出
	mixCmd := exec.CommandContext(ctx, "ffmpeg", "-y", "-i", rawVoice, "-i", "storage/assets/bgm.mp3", "-filter_complex", filter, "-c:a", "libmp3lame", outputPath)

	return mixCmd.Run()

}
