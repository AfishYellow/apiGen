package routers

import (
	"os"
	"strings"
)

func Start() {

	//add Handler to handler with prefix "router"
	AddHandler("router", func(item string, m map[string]string) {
		split := strings.Split(item, " ")
		m["router"] = split[0]
		m["action"] = split[1]
	})

	//currpath is the project path
	currpath, _ := os.Getwd()

	//toGen the router
	GenApiJson(currpath, "conf/server", "router")

	//	todo:insert to mysql
	/*	file, err := os.Open("conf/server/api.json")

		bytes, err := ioutil.ReadAll(file)

		if err != nil {
			println(err)
		}

		var apis []models.Api

		err = json.Unmarshal(bytes, &apis)
		if err != nil {
			fmt.Println("Marshal Error:",err.Error())
		}

		o := orm.NewOrm()

		for _, api := range apis {
			_, err := o.Insert(&api)
			if err != nil {
				fmt.Println("Insert Error:",err.Error())
			}
		}*/

}
