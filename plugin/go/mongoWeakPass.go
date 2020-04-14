package goplugin

import (
	"fmt"
	. "github.com/opensec-cn/kunpeng/config"
	"github.com/opensec-cn/kunpeng/plugin"
	"gopkg.in/mgo.v2"
	"strings"
	"time"
)

type mongoWeakPass struct {
	info   plugin.Plugin
	result []plugin.Plugin
}

func init() {
	plugin.Regist("mongodb", &mongoWeakPass{})
}
func (d *mongoWeakPass) Init() plugin.Plugin {
	d.info = plugin.Plugin{
		Name:    "MongoDB 未授权访问/弱口令",
		Remarks: "导致数据库敏感信息泄露，严重可导致服务器直接被入侵控制。",
		Level:   1,
		Type:    "WEAKPWD",
		Author:  "wolf",
		References: plugin.References{
			KPID: "KP-0007",
		},
	}
	return d.info
}
func (d *mongoWeakPass) GetResult() []plugin.Plugin {
	var result = d.result
	d.result = []plugin.Plugin{}
	return result
}
func (d *mongoWeakPass) Check(netloc string, meta plugin.TaskMeta) (b bool) {
	if strings.IndexAny(netloc, "http") == 0 {
		return
	}
	userList := []string{
		"admin",
	}
	session, err := mgo.Dial(netloc)
	if err == nil && session.Run("serverStatus", nil) == nil {
		result := d.info
		result.Request = fmt.Sprintf("mgo://%s/admin", netloc)
		result.Remarks = "未授权访问," + result.Remarks
		d.result = append(d.result, result)
		return true
	}
	for _, user := range userList {
		for _, pass := range meta.PassList {
			pass = strings.Replace(pass, "{user}", user, -1)
			dialInfo := &mgo.DialInfo{
				Addrs:     []string{netloc},
				Direct:    false,
				Timeout:   time.Second * time.Duration(Config.Timeout),
				Database:  "admin",
				Source:    "admin",
				Username:  user,
				Password:  pass,
				PoolLimit: 4096,
			}
			session, err := mgo.DialWithInfo(dialInfo)
			if err != nil {
				return
			}
			res, err := session.DatabaseNames()
			if err != nil {
				return
			}
			if res != nil {
				session.Close()
				result := d.info
				result.Request = fmt.Sprintf("mgo://%s:%s@%s/admin", user, pass, netloc)
				result.Remarks = fmt.Sprintf("弱口令：%s,%s,%s", user, pass, result.Remarks)
				d.result = append(d.result, result)
				b = true
				break
			}
		}
	}
	return b
}
