//14.创建deepseek.go,定义请求/响应体

package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Message 定义与大模型对话的单条消息格式（OpenAI 标准协议）
type Message struct {
	// Role 角色：system（设定AI人设）或 user（用户输入）
	Role string `json:"role"`
	// Content 具体发送给 AI 的文本内容/返回的内容
	Content string `json:"content"`
}

// ChatReq 发往 DeepSeek 官方接口的 POST 请求体
type ChatReq struct {
	// Model 指定调用的模型版本，通常用 "deepseek-chat"
	Model string `json:"model"`
	// Messages 上下文消息列表，按顺序传入人设与提问
	Messages []Message `json:"messages"`
}

// ChatResp 解析 DeepSeek 返回的 JSON 响应体
type ChatResp struct {
	// Choices 模型生成的候选回复列表，通常默认只有 1 条
	Choices []struct {
		// Message 模型真正回复的文案内容
		Message Message `json:"message"`
	} `json:"choices"`
}

//15.组装prompt,构建带Context超时的http请求
//传入上下文、密钥、主题，想deepseek请求200字口播文案,Generatescript-生成文本
	//ctx是30s的闹钟
func GenerateScript(ctx context.Context,apiKey string,topic string) (string,error){
	//组装prompt:设定人设，
	prompt := fmt.Sprintf("你是一位顶级宇宙物理纪录片的首席解说员。请围绕【%s】写一段200字左右充满哲学沉思、宏大深邃的中文解说词。严禁出现'家人们'、'大家好'、'今天聊聊'等网红口语，开篇直接从宏观宇宙视角展开叙事:", topic)
	reqBody := ChatReq{
		Model: "deepseek-flash",
		Messages: []Message{{Role:"user",Content: prompt}},
	}
	//将结构体序列化为JSON字节切片,Marshal 序列化
	data,_:= json.Marshal(reqBody)
	//绑定context超时监控，构建请求体
	httpReq,err := http.NewRequestWithContext(ctx,"POST","https://api.deepseek.com/chat/completions",bytes.NewBuffer(data))
	if err != nil{
		return "",err
	}
	//请求头设置
	httpReq.Header.Set("Content-Type","application/json")
	//注入Bearer令牌鉴权,
	httpReq.Header.Set("Authorization","Bearer " + apiKey)
	//使用go原生http客户端将请求发出
	resp,err := http.DefaultClient.Do(httpReq)
	if err != nil{
		return "",err
	}
	//defer 保证响应体网络流被彻底关闭,防止发生底层的socket连接泄露
	defer resp.Body.Close()
	//准备接大模型吐回来的数据
	var chatResp ChatResp
	//NewDecoder 边读网络流边反序列化成go结构体，decode解码
	if err := json.NewDecoder(resp.Body).Decode(&chatResp);err != nil{
		return "",err
	}
	//返回大模型生成的文案
	if len(chatResp.Choices) == 0{
		return "",fmt.Errorf("DeepSeek未返回有效文案")
	}
	return chatResp.Choices[0].Message.Content,nil
}


