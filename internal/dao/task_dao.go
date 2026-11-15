package dao

//Data Access Object 数据访问对象
//把跟mysql写的sql语句全部写在这
import (
	"echocraft/internal/model"
	"time"

	"gorm.io/gorm"
)

// TaskDAO 负责 tasks 表的所有数据库物理增删改查操作
//mysql <-> GORM <-> DAO
type TaskDAO struct{
	//持有GORM的底层连接操作对象
	db *gorm.DB
}

// NewTaskDAO 构造函数，把连接好的 db 注入进来完成初始化
func NewTaskDAO(db *gorm.DB) *TaskDAO {
	return &TaskDAO{db: db}
}

//Create 将一个新任务插入mysql数据库
func (d *TaskDAO) Create(task *model.Task) error{
	//执行insert into tasks，返回报错
	return d.db.Create(task).Error
}

func (d *TaskDAO) GetByTaskID(taskID string)(*model.Task,error){
	var task model.Task

	err := d.db.Where("task_id = ?",taskID).First(&task).Error
	return &task,err
}

func (d *TaskDAO) Update(task *model.Task,updates map[string]any) error{
	return d.db.Model(task).Updates(updates).Error
}

//返回 到底有没有改到行，数据库本身是否出错
func (d *TaskDAO) UpdateStatusCAS(TaskID string,from,to model.TaskStatus,updates map[string]any) (bool,error){
	if updates == nil{
		updates = map[string]any{}
	}
	updates["status"] = to

	result := d.db.Model(&model.Task{}).Where("task_id = ? AND status = ?",TaskID,from).Updates(updates)
	return result.RowsAffected > 0 ,result.Error
}

//按状态捞一批任务，启动恢复和后面的定时扫描
func (d *TaskDAO) ListByStatus(status model.TaskStatus,limit int) ([]model.Task,error){
	var tasks []model.Task	
	err := d.db.Where("status = ? ",status).Order("created_at asc").Limit(limit).Find(&tasks).Error
	return tasks,err
}

func (d *TaskDAO) ListStaleRunning(threshold time.Time,limit int) ([]model.Task,error){
	var tasks []model.Task
	err := d.db.Where("status = ? AND updated_at < ?",model.StatusRunning,threshold).Limit(limit).Find(&tasks).Error
	return tasks,err
}	