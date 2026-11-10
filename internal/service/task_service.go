package service

import (
	"context"
	"echocraft/internal/dao"
	"echocraft/internal/model"
	"echocraft/pkg/llm"
	"echocraft/pkg/tts"
	"errors"
	"fmt"
	"time"
)

// 业务调度中心，统领数据访问，任务队列，与外部api
type TaskService struct {
	dao      *dao.TaskDAO
	taskChan chan string
	apiKey   string
}

// 哨兵错误:队列满属于"预期内的失败",要用固定错误值,上层才能用 errors.Is 认出来
// 每次 fmt.Errorf 新建一个,外面就判断不了到底是不是队列满
var ErrQueueFull = errors.New("任务队列已满，请稍后重试")

//状态不对时，不能重试
var ErrorTaskNotRetryable = errors.New("任务当前状态不允许重试")

// 注入dao依赖，初始化缓冲队列
func NewTaskService(taskDao *dao.TaskDAO, apiKey string) *TaskService {
	s := &TaskService{
		dao:      taskDao,
		taskChan: make(chan string, 100),
		apiKey:   apiKey,
	}
	for i := 1; i <= 3; i++ {
		s.startWorker(i)
	}
	return s
}

// createTask 接收主题词，执行生成单号，调用dao落库，投递至队列，返回单号
func (s *TaskService) CreateTask(topic string) (string, error) {
	//生成一个单号
	taskID := fmt.Sprintf("task_%d", time.Now().UnixNano())
	//构造task实体
	task := model.Task{
		TaskID: taskID,
		Topic:  topic,
		Status: model.StatusPending,
	}
	//插入mysql
	if err := s.dao.Create(&task); err != nil {
		return "", err
	}

	select{
	case s.taskChan <- taskID:
		return taskID,nil
	default:
		//队列满了，标记错误
		_ = s.dao.Update(&task,map[string]any{
			"status": model.StatusFailed,
			"error_msg":ErrQueueFull.Error(),
		})
		return "",ErrQueueFull
	}
}

func (s *TaskService) GetTask(taskid string) (*model.Task, error) {
	return s.dao.GetByTaskID(taskid)
}

//传taskid重试
func (s *TaskService) RetryTask(taskid string) error {
	ok,err := s.dao.UpdateStatusCAS(taskid,model.StatusFailed,model.StatusPending,map[string]any{"err_msg":""})
	if err != nil{
		return err
	}
	if !ok{
		return ErrorTaskNotRetryable
	}
	select{
	case s.taskChan <- taskid:
		return nil
	default:
		//队列满了
		_,_ = s.dao.UpdateStatusCAS(taskid,model.StatusPending,model.StatusFailed,map[string]any{"err_msg":ErrQueueFull.Error()})
		return ErrQueueFull
	}
}

// 业务工作
func (s *TaskService) startWorker(workerID int){
	go func(){
		for taskID := range s.taskChan{
			fmt.Printf("[Worker %d]抢到任务:%s,开始处理...\n",workerID,taskID)
			s.processTask(workerID,taskID)
		}
	}()
}

func (s *TaskService) processTask(workerID int,taskID string){
	ctx,cancel := context.WithTimeout(context.Background(),time.Minute * 5)
	defer cancel()

	//抢任务,只有pending才能改为running
	//同一个任务被投递两次，用户连续点击两次重试，只有一个worker能抢到
	ok,err := s.dao.UpdateStatusCAS(taskID,model.StatusPending,model.StatusRunning,nil)
	if err != nil{
		fmt.Printf("[Worker %d]抢占任务 %s 失败 %v",workerID,taskID,err)
		return
	}
	if !ok {
		fmt.Printf("[Worker %d] %s 状态已变,跳过\n",workerID,taskID)
		return 
	}

	task,err := s.dao.GetByTaskID(taskID)
	if err != nil{
		fmt.Printf("[Worker %d]查询任务 %s 失败:%v \n",workerID,taskID,err)
		return
	}

	if task.ScriptContent == ""{
		//llmctx是一个闹钟，而调用llmcancel可以手动关闭闹钟
		llmCtx,llmCancel := context.WithTimeout(ctx,60*time.Second)
		script,err := llm.GenerateScript(llmCtx,s.apiKey,task.Topic)
		llmCancel()
		if err != nil{
			s.dao.Update(task,map[string]any{
				"status":model.StatusFailed,
				"err_msg":err.Error(),
			})
		}
		task.ScriptContent = script
		s.dao.Update(task,map[string]any{
			"status":model.StatusScriptDone,
			"script_content":script,
		})

		//tts合成
		ttsCtx,ttsCancel := context.WithTimeout(ctx,time.Second*120)
		audioPath := fmt.Sprintf("storage/%s.mp3",task.TaskID)
		err = tts.GenerateAudio(ttsCtx,script,audioPath)
		ttsCancel()
		if err != nil {
			s.dao.Update(task, map[string]any{"status": model.StatusFailed, "error_msg": err.Error()})
			return
		}

		s.dao.Update(task, map[string]any{"audio_path": audioPath, "status": model.StatusSuccess})
		fmt.Printf("[Worker %d]任务%s全流程完成,已落盘!\n", workerID, taskID)
	}
}