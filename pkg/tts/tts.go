package tts

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

//传入30s闹钟,文案，和保存路径，调用edge-tts生成mp3
//generateaudio = tts = text to speech
func GenerateAutdio(ctx context.Context,text string,outputPath string) error{
	//自动创建保存音频的父级目录,权限为0755
	if err := os.MkdirAll(filepath.Dir(outputPath),0755);err != nil{
		return err
	}

	//绑定超时，请操作系统的外部的edge-tts来配音
	cmd := exec.CommandContext(ctx,"edge-tts","--text",text,"--write-media",outputPath,"--voice","zh-CN-YunxiNeural")
	return cmd.Run()

}