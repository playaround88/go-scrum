package controllers

import (
	"github.com/astaxie/beego"
	"strings"
	"go-scrum/models"
	"net/http"
)

type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {
	c.Data["Website"] = "beego.me"
	c.Data["Email"] = "astaxie@gmail.com"
	c.TplName = "index.tpl"
}

const SessionKey string = "USER" 

type UserController struct {
	beego.Controller
}

//注册
func (c *UserController) Signup() {
	if c.Ctx.Request.Method == http.MethodGet {
		//
		c.TplName="signup.html"
	}else if c.Ctx.Request.Method == http.MethodPost {
		//TODO 获取参数，并验证

		pw1:=c.GetString("password1")
		pw2:=c.GetString("password2")
		if pw1!=pw2{
			c.Data["msg"]="两次密码不一致，请重新注册"
			c.TplName="signup.html"
			return
		}

		user := &models.User{
			Username:c.GetString("userName"),
			Password:pw1,
			Email:c.GetString("email"),
			Company:c.GetString("company"),
		}
		//持久化
		user.SaveOrUpdate()

		//跳转到登录页面
		c.Data["msg"]="注册成功，请登录"
		c.TplName="login.html"
	}
}


//登录
func (c *UserController) Login(){
	if c.Ctx.Request.Method == http.MethodGet {
		//返回登录页面
		c.TplName="login.html"
	}else if c.Ctx.Request.Method == http.MethodPost {
		//验证登录
		lg:=c.GetString("lg")
		password:=c.GetString("password")

		user := &models.User{}
		//验证是否是email
		if strings.Contains(lg,"@") {
			user.Email=lg
		}else{
			user.Username=lg
		}

		err := user.Load()
		if err!=nil {
			c.Data["msg"]="登录验证失败"
			c.TplName="login.html"
			return
		}

		//验证是否登录成功
		if password==user.Password {
			c.SetSession(SessionKey, user)
			c.TplName="main.html"
		}else {
			//返回登录页面
			c.Data["msg"]="登录验证失败"
			c.TplName="login.html"
		}
	}
}

//注销当前用户
func (c *UserController)Logout(){
	c.DestroySession()

	c.Data["msg"]="退出成功"
	c.TplName="login.html"
}


