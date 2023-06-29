package dashboard

import (
	"context"
	"github.com/gin-gonic/gin"
	v1 "github.com/shenyisyn/dbcore/pkg/apis/dbconfig/v1"
	"io/ioutil"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func isChildDeploy(cfg *v1.DbConfig, or corev1.ObjectReference, client client.Client) bool {
	//获取Deployment .命名规则是 dbcore-xxxx
	dep := &appv1.Deployment{}
	err := client.Get(context.Background(), types.NamespacedName{
		Name:      "dbcore" + cfg.Name,
		Namespace: cfg.Namespace,
	}, dep)
	if err != nil {
		return false
	}
	if or.UID == dep.UID {
		return true
	}
	return false

}

type AdminUi struct {
	r      *gin.Engine
	client client.Client
}

func NewAdminUi(c client.Client) *AdminUi {
	r := gin.New()
	r.StaticFS("/adminui", http.Dir("./adminui"))
	r.Use(errorHandler())
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	return &AdminUi{r: r, client: c}
}

func (this AdminUi) Start(c context.Context) error {
	this.r.GET("/events/:ns/:name", func(c *gin.Context) {
		var ns, name = c.Param("ns"), c.Param("name")

		// 本课时新增带啊
		cfg := v1.DbConfig{}
		Error(this.client.Get(c, types.NamespacedName{
			Name:      name,
			Namespace: ns,
		}, &cfg))
		list := &corev1.EventList{}
		Error(this.client.List(c, list, &client.ListOptions{Namespace: ns}))
		ret := []corev1.Event{}
		for _, e := range list.Items {
			//这是匹配 自定义资源 对应的 event
			//&& e.InvolvedObject.UID==cfg.UID
			if e.InvolvedObject.Name == name && e.InvolvedObject.UID == cfg.UID {
				ret = append(ret, e)
				continue
			}
			//代表判断，当前资源是否是dbconfig 创建出来的 deployment
			if isChildDeploy(&cfg, e.InvolvedObject, this.client) {
				ret = append(ret, e)
				continue
			}

		}
		c.JSON(200, ret)

	})

	this.r.GET("/configs", func(c *gin.Context) {
		list := v1.DbConfigList{}
		Error(this.client.List(c, &list))
		c.JSON(200, list.Items)
	})

	this.r.POST("/configs", func(c *gin.Context) {
		b, err := ioutil.ReadAll(c.Request.Body)
		Error(err)
		cfg := &v1.DbConfig{}
		Error(yaml.Unmarshal(b, cfg))
		if cfg.Namespace == "" {
			cfg.Namespace = "default"
		}
		Error(this.client.Create(c, cfg))
		c.JSON(200, gin.H{"message": "success"})
	})

	this.r.DELETE("/configs/:ns/:name", func(c *gin.Context) {
		cfg := v1.DbConfig{}
		err := this.client.Get(c, types.NamespacedName{
			Name:      c.Param("name"),
			Namespace: c.Param("ns"),
		}, &cfg)
		Error(err)
		Error(this.client.Delete(c, &cfg))
		c.JSON(200, gin.H{"message": "OK"})
	})

	return this.r.Run(":9003")
}
