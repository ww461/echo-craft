package dao

//Data Access Object 数据访问对象
//把跟mysql写的sql语句全部写在这
import(
	"echocraft/internal/model"
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