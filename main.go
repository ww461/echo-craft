package main

import (
	"echocraft/internal/api"
	"echocraft/internal/dao"
	"echocraft/internal/model"
	"echocraft/internal/service"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)


func main() {
	_= godotenv.Load()
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if(apiKey == ""){
		panic("缺失API_KEY配置,请检查.env文件!")
	}
	
	dsn := "root:123456@tcp(127.0.0.1:3306)/echocraft?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("连接数据库失败：" + err.Error())
	}
	_ = db.AutoMigrate(&model.Task{})

	taskDAO := dao.NewTaskDAO(db)
	TaskService := service.NewTaskService(taskDAO,apiKey)
	TaskHandler := api.NewTaskHandler(TaskService)

	r := gin.Default()
	TaskHandler.RegisterRoutes(r)

	r.Run(":9090")
}
