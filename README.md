# EchoCraft: AIGC 异步任务调度与短播客生成引擎

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Gin Framework](https://img.shields.io/badge/Gin-v1.10.0-008ECF?style=flat&logo=gin)](https://gin-gonic.com/)
[![GORM](https://img.shields.io/badge/GORM-v1.25.0-blue?style=flat)](https://gorm.io/)
[![MySQL](https://img.shields.io/badge/MySQL-8.0-4479A1?style=flat&logo=mysql)](https://www.mysql.com/)
[![DeepSeek](https://img.shields.io/badge/LLM-DeepSeek-0066FF?style=flat)](https://deepseek.com/)

> **EchoCraft** 是一个面向 AIGC 长耗时场景的高性能、轻量级异步任务编排与短播客生成服务。  
> 客户端提交一个生成主题后，服务端在 **10ms 内极速返回任务单号**；后台基于 **Go 缓冲 Channel 与常驻 Worker 池** 实现削峰消费，全链路串联大语言模型（DeepSeek）文案生成与微软 Edge-TTS 语音合成，最终交付单文件 `.mp3` 播客成果。
