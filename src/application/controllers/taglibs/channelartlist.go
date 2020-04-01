package taglibs

import (
	"fmt"
	"github.com/CloudyKit/jet"
	"github.com/go-xorm/xorm"
	"github.com/xiusin/pine"
	"github.com/xiusin/pinecms/src/application/controllers"
	"github.com/xiusin/pinecms/src/application/models"
	"github.com/xiusin/pinecms/src/application/models/tables"
	"reflect"
	"strconv"
	"strings"
)

func ChannelArtList(args jet.Arguments) reflect.Value {
	defer func() {
		if err := recover(); err != nil {
			pine.Logger().Error("ChannelArtList Failed", err)
		}
	}()
	catid := args.Get(0)
	row := getInt(args.Get(1))
	if row == 0 {
		row = 10
	}
	// 是否包含是否有下级的属性
	var ids []string
	str := catid.Type().String()
	if strings.HasPrefix(str, "float") {
		ids = append(ids, strconv.Itoa(int(catid.Float())))
	} else if strings.Contains(str, "int") {
		ids = append(ids, strconv.Itoa(int(catid.Int())))
	} else {
		ids = strings.Split(catid.String(), ",")
	}
	sess := pine.Make(controllers.ServiceXorm).(*xorm.Engine).Table(&tables.Category{})
	var categories []tables.Category
		err := sess.In("parentid", ids).Where("ismenu = ?", 1).Limit(int(row)).Asc("listorder").Find(&categories)
	if err != nil {
		panic(err)
	}
	var catids = []int64{}
	m := models.NewCategoryModel()
	for k, v := range categories {
		if v.Type != 2 {
			categories[k].Url = fmt.Sprintf("/%s/", m.GetUrlPrefix(v.Catid))
		}
		catids = append(catids, v.Catid)
	}

	// 判断是否有下级菜单
	son := args.Get(2).Bool()
	if son {
		sess := pine.Make(controllers.ServiceXorm).(*xorm.Engine).
			Table(&tables.Category{}).
			In("parentid", catids).GroupBy("parentid").Select("parentid,count(*) as total").Where("ismenu = ?", 1)
		rest, _ := sess.QueryString()
		var kvPairs = map[string]string{}
		for _, v := range rest {
			kvPairs[v["parentid"]] = v["total"]
			//todo category Son级联查出. 一次性暴露所有的分类, 无需再次使用channel调用
		}
		for k, v := range categories {
			val, exists := kvPairs[strconv.Itoa(int(v.Catid))]
			if exists && len(val) > 0 && val != "0" {
				categories[k].HasSon = true
			}
			if v.Type != 2 {
				categories[k].Url = fmt.Sprintf("/%s/", m.GetUrlPrefix(v.Catid))
			}
		}
	}
	return reflect.ValueOf(categories)
}
