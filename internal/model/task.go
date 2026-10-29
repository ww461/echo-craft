package model

import "time"

// 2.定义任务状态常量
type TaskStatus string
const (
	// StatusPending 任务已入库排队
	StatusPending TaskStatus = "PENDING"
	// StatusScriptDone AI 文案已生成入库
	StatusScriptDone TaskStatus = "SCRIPT_DONE"
	// StatusSuccess 音频合成成功，流程结束
	StatusSuccess TaskStatus = "SUCCESS"
	// StatusFailed 执行中途报错失败
	StatusFailed TaskStatus = "FAILED"
)


// 3.定义任务模型Task结构体
// task 结构体对应数据库的一行记录，也是贯穿整个调度流程的核心数据载体
type Task struct {
	// ID 数据库自增主键
	ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	// TaskID 外部暴露的业务单号，不直接暴露自增 ID 保证安全
	TaskID string `gorm:"type:varchar(64);uniqueIndex;not null" json:"task_id"`
	// Topic 用户输入的生成主题（例如“如何学好Go语言”）
	Topic string `gorm:"type:varchar(255);not null" json:"topic"`
	// Status 任务当前状态，取值就是刚才定义的 4 个常量之一
	Status TaskStatus `gorm:"type:varchar(32);index;not null" json:"status"`
	// ScriptContent AI 生成的文案，断点续跑时发现它不为空就直接跳过大模型
	ScriptContent string `gorm:"type:text" json:"script_content"`
	// AudioPath 本地合成好的 .mp3 文件的保存路径
	AudioPath string `gorm:"type:varchar(512)" json:"audio_path"`
	// ErrorMsg 任何一步执行失败时记录的报错信息
	ErrorMsg string `gorm:"type:varchar(512)" json:"error_msg"`
	// CreatedAt 记录任务创建的时间
	CreatedAt time.Time `json:"created_at"`
}


// 5.定义请求参数结构体
// CreateTaskReq 专门用来接收客户端 POST 提交的 JSON 参数
type CreateTaskReq struct {
	// Topic 用户输入的生成主题，binding:"required" 声明为必填项
	//binding数据绑定，把HTTP请求的数据自动解析并填充到定义的结构体
	Topic string `json:"topic" binding:"required"`
}