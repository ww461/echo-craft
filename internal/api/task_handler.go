package api

import (
	"echocraft/internal/model"
	"echocraft/internal/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

//http控制层，调用service
type TaskHandler struct{
	service *service.TaskService
}


//完成依赖注入
func NewTaskHandler(service *service.TaskService) *TaskHandler{
	return &TaskHandler{service: service}
}


//处理post请求
func (h *TaskHandler) CreateTask(c *gin.Context){
	var req model.CreateTaskReq

	if err := c.ShouldBindJSON(&req);err != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":"参数校验失败,topic为必填项"})
		return
	}
	taskID,err := h.service.CreateTask(req.Topic)
	if err != nil{
		//队伍满了，服务暂时忙，返回503
		if errors.Is(err,service.ErrQueueFull){
			c.JSON(http.StatusServiceUnavailable,gin.H{"error":err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError,gin.H{"error":"创建任务失败:" + err.Error()})
		return
	}
	c.JSON(http.StatusOK,gin.H{"task_id":taskID,"status":model.StatusPending})
}

func (h *TaskHandler) GetTask(c *gin.Context){

	taskID := c.Param("task_id")

	task,err := h.service.GetTask(taskID)
	if err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"任务单号不存在"})
		return
	}
	c.JSON(200,gin.H{"code":200,"message":"success","data":task})
}

func (h *TaskHandler) RetryTask(c *gin.Context){
	taskID := c.Param("task_id")

	//满了重试报错
	if err := h.service.RetryTask(taskID);err != nil{
		if errors.Is(err,service.ErrorTaskNotRetryable){
			c.JSON(http.StatusConflict,gin.H{"error" : err.Error()})
			return
		}
		if errors.Is(err,service.ErrQueueFull){
			c.JSON(http.StatusServiceUnavailable,gin.H{"error":err.Error()})
			return
		}
	}
	
	c.JSON(http.StatusOK,gin.H{"message":"重试任务已进入后台排队","task_id":taskID})
}

func (h *TaskHandler) RegisterRoutes(r *gin.Engine){
	r.POST("/api/v1/tasks",h.CreateTask)
	r.GET("api/v1/tasks/:task_id",h.GetTask)
	r.POST("api/v1/tasks/:task_id/retry",h.RetryTask)
}