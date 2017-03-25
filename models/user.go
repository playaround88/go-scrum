package models

import (
	"fmt"
	"github.com/pkg/errors"
	"strconv"
)

//User 系统用户
//redis存储结构
//user:seq - string Id序列值
//user:[Id] - Hash 对象存储
//contactor:[User_Id] - list 联系人
//登录映射
//un2id - hash 用户名到id的映射
//email2id - hash Email到id的映射
type User struct {
	Id int64
	Username string
	Password string
	Email    string
	Company string

	Contactors []User
}

const (
	USER_PREFIX="user:"
	USER_SEQ=USER_PREFIX+"seq"
	UN2ID="un2id"
	EMAIL2ID="email2id"

	CONTACTOR="contactor:"
)

//添加用户或者Id非零更新用户
func (u *User) SaveOrUpdate() error{
	//判断Id是否为空
	if u.Id <= 0{
		ic:=client.Incr(USER_SEQ)
		u.Id=ic.Val()
	}

	//使用事务管道，节省网络开支
	pipeline:=client.TxPipeline()
	//保存对象
	pipeline.HSet(USER_PREFIX+fmt.Sprintf("%d",u.Id),"UserName",u.Username)
	pipeline.HSet(USER_PREFIX+fmt.Sprintf("%d",u.Id),"Password",u.Password)
	pipeline.HSet(USER_PREFIX+fmt.Sprintf("%d",u.Id),"Email",u.Email)
	pipeline.HSet(USER_PREFIX+fmt.Sprintf("%d",u.Id),"Company",u.Company)

	//保存un2id、email2id映射关系
	pipeline.HSet(UN2ID,u.Username,u.Id)
	pipeline.HSet(EMAIL2ID,u.Email,u.Id)

	_,err:=pipeline.Exec()

	return err
}

func (u *User) Del() error{
	if u.Id <= 0 {
		return errors.New("Id值异常，未执行删除操作")
	}
	//注意这里要先加载，否则不能保证username email为空
	u.Load()

	//
	pipeline := client.TxPipeline()

	pipeline.Del(USER_PREFIX+fmt.Sprintf("%d",u.Id))
	//删除映射
	pipeline.HDel(UN2ID, u.Username)
	pipeline.HDel(EMAIL2ID,u.Email)

	_,err:=pipeline.Exec()

	return err
}

func (u *User) Load() error{
	var user User
	var err error

	//加载用户
	if u.Id >0 {
		user,err = LoadUserById(u.Id)
	}else if u.Username != "" {
		user,err = LoadUserByUN(u.Username)
	}else if u.Email != "" {
		user,err = LoadUserByEmail(u.Email)
	}

	if err!=nil{
		return err
	}

	*u = user
	return nil
}

//通过Email查询User
func LoadUserById(id int64) (User,error){
	ssm:=client.HGetAll(USER_PREFIX+fmt.Sprintf("%d",id))

	if ssm.Err() != nil {
		return User{},ssm.Err()
	}
	//回填Id
	u,err := mapUser(ssm.Val())
	u.Id=id
	return u,err
}

//通过userName查询User
func LoadUserByUN(userName string) (User,error) {
	sc:=client.HGet(UN2ID,userName)
	ssm := client.HGetAll(USER_PREFIX+sc.Val())

	if ssm.Err() != nil {
		return User{},ssm.Err()
	}

	u,err:=mapUser(ssm.Val())
	//回填Id
	i,err := strconv.ParseInt(sc.Val(),10,64)
	if err!=nil {
		return u, err
	}
	u.Id=i

	return u,err
}

//通过Email查询User
func LoadUserByEmail(email string) (User,error){
	sc:=client.HGet(EMAIL2ID,email)
	ssm := client.HGetAll(USER_PREFIX+sc.Val())

	if ssm.Err() != nil {
		return User{},ssm.Err()
	}

	u,err:=mapUser(ssm.Val())
	//回填Id
	i,err := strconv.ParseInt(sc.Val(),10,64)
	if err!=nil {
		return u, err
	}
	u.Id=i

	return u,err
}

//映射查询结果
func mapUser(m map[string]string) (User,error){
	u := User{}

	u.Username=m["UserName"]
	u.Password=m["Password"]
	u.Email=m["Email"]
	u.Company=m["Company"]

	return u,nil
}
