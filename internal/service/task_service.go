package service

import (
	"context"
	"echocraft/internal/dao"
	"echocraft/internal/model"
	"echocraft/pkg/llm"
	"echocraft/pkg/tts"
	"fmt"
	"time"
)

//业务调度中心，统领数据访问，任务队列，与外部api
type TaskService struct{
	dao *dao.TaskDAO
	taskChan chan string
	apiKey string
}


//注入dao依赖，初始化缓冲队列
func NewTaskService(taskDao *dao.TaskDAO,apiKey string) *TaskService{
	s := &TaskService{
		dao: taskDao,
		taskChan: make(chan string,100),
		apiKey: apiKey,
	}
	for i := 1;i <= 3;i++{
		s.startWorker(i)
	}
	return s
}

//createTask 接收主题词，执行生成单号，调用dao落库，投递至队列，返回单号
func(s *TaskService) CreateTask(topic string) (string,error){
	//生成一个单号
	taskID := fmt.Sprintf("task_%d",time.Now().UnixNano())
	//构造task实体
	task := model.Task{
		TaskID: taskID,
		Topic: topic,
		Status: model.StatusPending,
	}
	//插入mysql
	if err := s.dao.Create(&task); err != nil{
		return "",err
	}

	s.taskChan <- taskID

	return taskID,nil
}

//
func(s *TaskService) GetTask(taskid string) (*model.Task,error){
	return s.dao.GetByTaskID(taskid)
}

func (s *TaskService)RetryTask(taskid string) error{
	s.taskChan <- taskid

	return nil
}

//业务工作
func(s *TaskService) startWorker(workerID int){
	go func(){
		for taskID := range s.taskChan{
			fmt.Printf("[Worker %d]抢到任务:%s，开始处理...\n",workerID,taskID)
			//ctx是30s的超时context
			ctx,cancel := context.WithTimeout(context.Background(),30*time.Second)

			//拿到task本体
			task,err := s.dao.GetByTaskID(taskID)
			if err != nil{
				fmt.Printf("[Worker %d]查询任务 %s 失败：%v\n",workerID,taskID,err)
				cancel()
				continue
			}
			if task.ScriptContent == ""{
				script,err := llm.GenerateScript(ctx,s.apiKey,task.Topic)
				if err != nil{
					s.dao.Update(task,map[string]any{
						"status": model.StatusFailed,
						"error_msg": err.Error(),
					})
					cancel()
					continue
				}
				task.ScriptContent = script
				s.dao.Update(task,map[string]any{
					"script_content": script,
					"status":model.StatusScriptDone,
				})
			}
			audioPath := fmt.Sprintf("storage/%s.mp3",task.TaskID)
			if err:= tts.GenerateAutdio(ctx,task.ScriptContent,audioPath);err != nil{
				s.dao.Update(task,map[string]any{
					"status":model.StatusFailed,
					"error_msg":err.Error(),
				})
				cancel()
				continue
			}
			s.dao.Update(task,map[string]any{
				"audio_path": audioPath,
				"status":model.StatusSuccess,
			})

			fmt.Printf("[Worker %d]任务%s全流程完成,已落盘!\n",workerID,taskID)
			cancel()
		}
	}()
}