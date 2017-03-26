package models

import (
	"time"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"log"
)

//Project 项目
//redis存储结构
//project:[Id] - hash 对象存储
//project:seq - string Id序列值
//user_projects:[user_id] - list 用户项目列表
//user_a_projects:[user_id] -list 用户已归档项目列表
type Project struct {
	Id int64
	Name string
	Desc string
	Creator int64
	CreateTime time.Time

	//状态，是否归档
	State string
}

const (
	PROJECT_PREFIX string = "project:"
	PROJECT_SEQ string = PROJECT_PREFIX + "seq"
	USER_PROJECT string = "user_project:"
	USER_A_PROJECT string = "user_a_project:"
)

func (p *Project) SaveOrUpdate() error{
	//如果Id不存在，则为新添加
	if p.Id <= 0 {
		i:=client.Incr(PROJECT_SEQ)
		p.Id=i.Val()
	}
	//pipeline，节省网络开销
	pipeline := client.Pipeline()

	pipeline.HSet(PROJECT_PREFIX+fmt.Sprintf("%d",p.Id), "Name", p.Name)
	pipeline.HSet(PROJECT_PREFIX+fmt.Sprintf("%d",p.Id), "Desc", p.Desc)
	pipeline.HSet(PROJECT_PREFIX+fmt.Sprintf("%d",p.Id),
		"Creator",
		fmt.Sprintf("%d",p.Creator))
	pipeline.HSet(PROJECT_PREFIX+fmt.Sprintf("%d",p.Id),
		"CreateTime",
		fmt.Sprintf("%d",p.CreateTime.Unix()))

	if p.State!="" {
		pipeline.HSet(PROJECT_PREFIX+fmt.Sprintf("%d",p.Id), "State", p.State)
	}

	//关联关系
	pipeline.LPushX(USER_PROJECT+fmt.Sprintf("%d",p.Creator),fmt.Sprintf("%d",p.Id))

	_,err:=pipeline.Exec()
	return err

}

func (p *Project) Del() error {
	if p.Id <= 0 {
		return errors.New("项目编码异常，未做删除操作")
	}

	//先加载，否则可能不能正常删除关联关系
	if p.Creator <= 0 {
		return errors.New("项目数据有异常，请检查Creator字段数据")
	}

	pipeline := client.TxPipeline()

	pipeline.Del(PROJECT_PREFIX+fmt.Sprintf("%d"))
	//删除关联关系
	pipeline.LRem(USER_PROJECT+fmt.Sprintf("%d",p.Creator),1,p.Id)
	pipeline.LRem(USER_A_PROJECT+fmt.Sprintf("%d",p.Creator),1,p.Id)

	_,err := pipeline.Exec()

	return err
}

func (p *Project) Load() error{
	if p.Id <= 0 {
		return errors.New("项目编码异常，未做查询操作")
	}

	ssm:=client.HGetAll(PROJECT_PREFIX+fmt.Sprintf("%d"))
	if ssm.Err() {
		return ssm.Err()
	}

	m:=ssm.Val()

	return mapProject(p,m)
}

func mapProject(p *Project, m map[string]string) error {
	p.Name=m["Name"]
	p.Desc=m["Desc"]
	//创建人
	uid,err:=strconv.ParseInt(m["Creator"],10,64)
	if err != nil {
		return err
	}
	p.Creator=uid
	//时间转换
	int64Time,err := strconv.ParseInt(m["CreateTime"],10,64)
	if err != nil {
		return err
	}
	p.CreateTime = time.Unix(int64Time,0)
	//状态
	p.State=m["State"]

	return nil
}

//Archive 项目归档
func (p *Project) Archive() error{
	if p.State=="archived" {
		return nil
	}
	//
	if p.Id <= 0 || p.Creator <= 0 {
		return errors.New("项目数据有异常，请检查Id和Creator字段数据")
	}

	p.State="archived"

	pipeline:=client.TxPipeline()

	pipeline.LRem(USER_PROJECT+fmt.Sprintf("%d",p.Creator),1,p.Id)
	pipeline.LPushX(USER_A_PROJECT+fmt.Sprintf("%d",p.Creator), p.Id)

	_,err:=pipeline.Exec()

	return err
}

func QueryProjectsByUser(userId int64) []Project {
	uid:=fmt.Sprintf("%d",userId)

	ic:=client.LLen(USER_A_PROJECT+uid)
	len := ic.Val()

	ssc:=client.LRange(USER_A_PROJECT+uid,0,len)
	projectIds:=ssc.Val()

	var projects []Project

	for _, projectId := range projectIds {
		//初始化project
		pid,err := strconv.ParseInt(projectId,10,64)
		if err != nil {
			log.Println(err)
		}
		project := Project{Id:pid}
		//加载
		project.Load()
		//添加到slice
		projects=append(projects,project)
	}

	return projects
}

func QueryAProjectsByUser(userId string) []Project{
	uid:=fmt.Sprintf("%d",userId)

	ic:=client.LLen(USER_A_PROJECT+uid)
	len := ic.Val()

	ssc:=client.LRange(USER_A_PROJECT+uid,0,len)
	projectIds:=ssc.Val()

	var projects []Project

	for _, projectId := range projectIds {
		//初始化project
		pid,err := strconv.ParseInt(projectId,10,64)
		if err != nil {
			log.Println(err)
		}
		project := Project{Id:pid}
		//加载
		project.Load()
		//添加到slice
		projects=append(projects,project)
	}

	return projects
}