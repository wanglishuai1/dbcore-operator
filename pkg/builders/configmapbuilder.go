package builders

import (
	"bytes"
	"context"
	configv1 "github.com/shenyisyn/dbcore/pkg/apis/dbconfig/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"log"
	"text/template"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ConfigMapBuilder struct {
	cm     *corev1.ConfigMap
	config *configv1.DbConfig //新增属性 。保存 config 对象
	client.Client
	DataKey string
}

// 构建 cm 创建器
func NewConfigMapBuilder(config *configv1.DbConfig, client client.Client) (*ConfigMapBuilder, error) {
	cm := &corev1.ConfigMap{}
	err := client.Get(context.Background(), types.NamespacedName{
		Namespace: config.Namespace, Name: fmtName(config.Name), //这里做了改动
	}, cm)
	if err != nil { //没取到
		//只需要 赋值 name 和namespace data不管，在apply 函数中处理
		cm.Name, cm.Namespace = fmtName(config.Name), config.Namespace
		cm.Data = make(map[string]string) //搞一个空的map即可
	}
	return &ConfigMapBuilder{cm: cm, Client: client, config: config}, nil
}

// 同步属性
func (this *ConfigMapBuilder) setOwner() *ConfigMapBuilder {
	this.cm.OwnerReferences = append(this.cm.OwnerReferences,
		metav1.OwnerReference{
			APIVersion: this.config.APIVersion,
			Kind:       this.config.Kind,
			Name:       this.config.Name,
			UID:        this.config.UID,
		})
	return this
}
func (this *ConfigMapBuilder) parseKey() *ConfigMapBuilder {
	if appData, ok := this.cm.Data[configMapKey]; ok {
		this.DataKey = Md5(appData)
		return this
	}
	this.DataKey = ""
	return this
}

const configMapKey = "app.yml"

func (this *ConfigMapBuilder) apply() *ConfigMapBuilder {
	//下节课来写
	tpl, err := template.New("appyaml").
		Delims("[[", "]]").
		Parse(cmtpl)
	if err != nil {
		log.Println(err)
		return this
	}
	var tplRet bytes.Buffer
	err = tpl.Execute(&tplRet, this.config.Spec)
	if err != nil {
		log.Println(err)
		return this
	}
	this.cm.Data[configMapKey] = tplRet.String()
	return this
}

// 构建出  ConfigMap  ..有可能是新建， 有可能是update
func (this *ConfigMapBuilder) Build(ctx context.Context) error {
	if this.cm.CreationTimestamp.IsZero() {
		this.apply().setOwner().parseKey() //同步  所需要的属性 如 副本数 , 并且设置OwnerReferences
		err := this.Create(ctx, this.cm)
		if err != nil {
			return err
		}
	} else {
		patch := client.MergeFrom(this.cm.DeepCopy())
		this.apply().parseKey() //同步  所需要的属性 如 副本数
		err := this.Patch(ctx, this.cm, patch)
		if err != nil {
			return err
		}

	}
	//更新 是课后作业
	return nil

}
