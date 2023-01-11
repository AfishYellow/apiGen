// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"apiGen/controllers"
	"github.com/astaxie/beego"
)

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
