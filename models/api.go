package models

// Api api
type Api struct {
	Id          int64  `orm:"pk;auto"json:"id"`
	Status      int64  `orm:"size(1);default(0);description(0是启用 1是删除)"json:"status"`
	Add_time    int64  `orm:"size(20);description(创建时间)"json:"add_time"`
	Update_time int64  `orm:"size(20);description(更新时间)"json:"update_time"`
	Name        string `orm:"size(255);description(名字)"json:"name""`
	Path        string `orm:"size(100);description(路径)"json:"path""`
	Type        string `orm:"size(100);description(接口类型)"json:"type""`
	Action      string `orm:"size(100);description(请求方式)"json:"action""`
	Remark      string `orm:"size(255);description(备注)"json:"remark""`
}
