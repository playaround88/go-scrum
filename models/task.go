package models

import (
	"time"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
)

//Task 任务条目
//redis存储结构
//task:[id] - hash
//task:seq - string Id序列值
//lv_task:[lvId] - set listview与任务列表
//lane_task:[lvId] - set 泳道任务列表
type Task struct {
	Id int64
	Name string
	Desc string
	//关联
	ListviewId int64
	LaneId int64

	//时间
	CreateTime time.Time
	EndTime time.Time
	//附件Ids（comma seperate）
	Attachement string

	//责任人
	Member []User
}

const (
	TASK_PRFIX = "task:"
	TASK_SEQ = TASK_PRFIX+"seq"
	LV_TASK = "lv_task:"
	LANE_TASK = "lane_task:"
)

func (t *Task) SaveOrUpdate() error {
	if t.Id <= 0 {
		ic := client.Incr(TASK_SEQ)
		if ic.Err() != nil {
			return ic.Err()
		}
		t.Id = ic.Val()
	}

	pipeline := client.TxPipeline()

	pipeline.HSet(TASK_PRFIX + fmt.Sprintf("%d",t.Id),"Name", t.Name)
	pipeline.HSet(TASK_PRFIX + fmt.Sprintf("%d",t.Id),"Desc", t.Desc)
	pipeline.HSet(TASK_PRFIX + fmt.Sprintf("%d",t.Id),"ListviewId", fmt.Sprintf("%d",t.ListviewId))
	pipeline.HSet(TASK_PRFIX + fmt.Sprintf("%d",t.Id),"LaneId", fmt.Sprintf("%d",t.LaneId))
	pipeline.HSet(TASK_PRFIX + fmt.Sprintf("%d",t.Id),"CreateTime", fmt.Sprintf("%d", t.CreateTime.Unix()))
	pipeline.HSet(TASK_PRFIX + fmt.Sprintf("%d",t.Id),"EndTime", fmt.Sprintf("%d",t.EndTime.Unix()))
	pipeline.HSet(TASK_PRFIX + fmt.Sprintf("%d",t.Id),"Attachement", t.Attachement)
	//关联关系
	pipeline.SAdd(LV_TASK+fmt.Sprintf("%d",t.ListviewId), fmt.Sprintf("%d",t.Id))
	pipeline.SAdd(LANE_TASK+fmt.Sprintf("%d",t.LaneId), fmt.Sprintf("%d",t.Id))

	_,err := pipeline.Exec()

	return err

}

func (t *Task) Del() error{
	if t.Id <= 0 || t.ListviewId <= 0 || t.LaneId <= 0 {
		return errors.New("数据异常，未操作，check Id/ListviewId/LaneId field")
	}
	pipeline := client.TxPipeline()
	pipeline.Del(TASK_PRFIX + fmt.Sprintf("%d",t.Id))
	pipeline.SRem(LV_TASK+fmt.Sprintf("%d",t.ListviewId),fmt.Sprintf("%d",t.Id))
	pipeline.SRem(LANE_TASK+fmt.Sprintf("%d",t.ListviewId),fmt.Sprintf("%d",t.Id))

	_,err := pipeline.Exec()
	return err
}

func (t *Task) Load() error {
	if t.Id <= 0 {
		return errors.New("数据异常，未操作，check Id field")
	}

	ssc := client.HGetAll(TASK_PRFIX + fmt.Sprintf("%d",t.Id))
	if ssc.Err() != nil {
		return ssc.Err()
	}
	m := ssc.Val()

	return mapTask(t, m)
}

func mapTask(t *Task, m map[string]string) error {
	t.Name = m["Name"]
	t.Desc = m["Desc"]

	lvId,err := strconv.ParseInt(m["ListviewId"],10,64)
	if err != nil {
		return err
	}
	t.ListviewId = lvId

	lId,err := strconv.ParseInt(m["LaneId"],10,64)
	if err != nil {
		return err
	}
	t.LaneId = lId

	csec,err := strconv.ParseInt(m["CreateTime"], 10, 64)
	if err != nil {
		return err
	}
	t.CreateTime=time.Unix(csec,0)

	esec,err := strconv.ParseInt(m["EndTime"],10,64)
	if err != nil {
		return err
	}
	t.EndTime = time.Unix(esec,0)

	t.Attachement = m["Attachement"]

	return err
}

func QueryTaskByLV(lvId int64) ([]Task,error) {
	var tasks []Task
	ssc :=client.SMembers(LV_TASK+fmt.Sprintf("%d",lvId))
	if ssc.Err() != nil {
		return tasks,ssc.Err()
	}

	taskIds := ssc.Val()
	for _,tid := range taskIds {
		task := Task{}

		taskId,err := strconv.ParseInt(tid,10,64)
		if err != nil {
			return tasks,err
		}
		task.Id=taskId

		task.Load()

		tasks = append(tasks,task)
	}

	return tasks,nil
}

func QueryTaskByLane(laneId int64) ([]Task,error) {
	var tasks []Task
	ssc :=client.SMembers(LANE_TASK+fmt.Sprintf("%d",laneId))
	if ssc.Err() != nil {
		return tasks,ssc.Err()
	}

	taskIds := ssc.Val()
	for _,tid := range taskIds {
		task := Task{}

		taskId,err := strconv.ParseInt(tid,10,64)
		if err != nil {
			return tasks,err
		}
		task.Id=taskId

		task.Load()

		tasks = append(tasks,task)
	}

	return tasks,nil
}