该项目初衷是为了可以生成通过注解可以生成 api json或api结构体

该项目在beego的api上面测试成功

将以@开头的注解转换成对应的json

```go
// @Title Login
// @Description Logs user into the system
// @Param  username      query  string true      "The username for login"
// @Param  password      query  string true      "The password for login"
// @Success 200 {string} login success
// @Failure 403 user not exist
// @router /login [get]
func (u *UserController) Login(i int) {
   username := u.GetString("username")
   password := u.GetString("password")
   if models.Login(username, password) {
      u.Data["json"] = "login success"
   } else {
      u.Data["json"] = "user not exist"
   }
   u.ServeJSON()
}
```

转化后的格式如下:

```json
{
  "Description": "Logs user into the system",
  "Failure": "403 user not exist",
  "Param": "password  query  string true  \"The password for login\"",
  "Success": "200 {string} login success",
  "Title": "Login",
  "action": "[get]",
  "router": "/login"
},
```



### 使用方法:

#### 1.将routers/gen.go复制到项目routers文件夹下

#### 2.在routers/router.go中设置好命名空间:

/v1/object

/v1/user

```go
func init() {
   ns := beego.NewNamespace("/v1",
      beego.NSNamespace("/object",
         beego.NSInclude(
            &controllers.ObjectController{},
         ),
      ),
      beego.NSNamespace("/user",
         beego.NSInclude(
            &controllers.UserController{},
         ),
      ),
   )
   beego.AddNamespace(ns)

   /*
      // 如果使用的不是beego框架，也可以使用以下的代码进行命名空间的设置
      NewNamespace("/v1",
         NSNamespace("/object",
            NSInclude(
               &controllers.ObjectController{},
            ),
         ),
         NSNamespace("/user",
            NSInclude(
               &controllers.UserController{},
            ),
         ),
      )
   */
}
```

#### 3.开始生成

```go
func main() {


//添加处理器对指定的前辍  @router 
//@router /login [get] => map{ "router":"/login","action"="[get]" }

   //add Handler to handler with prefix "router"
   
   routers.AddHandler("router", func(item string, m map[string]string) {
      split := strings.Split(item, " ")
      m["router"] = split[0]
      m["action"] = split[1]
   })

   //currpath is the project path
   currpath, _ := os.Getwd()

    //第三个参数需要手动标识需添加命名空间作为前辍的字段
    //map{ "router":"/login" } =>map{ "router":"/v1/user/login"}
   //toGen the router
   routers.GenApiJson(currpath, "conf/server", "router")

}
```

即可生成  /conf/server/api.json 