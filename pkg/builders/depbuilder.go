package builders

import (
	"bytes"
	"context"
	"fmt"
	dbconfigv1 "github.com/shenyisyn/dbcore/pkg/apis/dbconfig/v1"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"text/template"
)

type DeployBuilder struct {
	deploy    *appv1.Deployment
	config    *dbconfigv1.DbConfig
	cmBuilder *ConfigMapBuilder // 关联  对象
	client.Client
}

func fmtName(name string) string {
	return "dbcore" + name
}

func NewDeployBuilder(config dbconfigv1.DbConfig, client client.Client) (*DeployBuilder, error) {
	dep := &appv1.Deployment{}
	err := client.Get(context.Background(), types.NamespacedName{Namespace: config.Namespace, Name: fmtName(config.Name)}, dep)
	if err != nil { //没取到，就新建一个
		dep.Name, dep.Namespace = config.Name, config.Namespace
		tmp, err := template.New("deploy").Parse(deptmp)
		if err != nil {
			return nil, err
		}
		var doc bytes.Buffer
		err = tmp.Execute(&doc, dep)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(doc.Bytes(), dep)
		if err != nil {
			return nil, err
		}
		dep.Spec.Template.Annotations = make(map[string]string) //搞一个空的map即可
	}
	//configmap 构建
	cmBuilder, err := NewConfigMapBuilder(&config, client)
	if err != nil {
		fmt.Println("cm error:", err)
		return nil, err
	}
	return &DeployBuilder{
		deploy:    dep,
		Client:    client,
		config:    &config,
		cmBuilder: cmBuilder,
	}, nil
}

const CMAnnotation = "dbcore.config/md5"

// 更新configmap时，更新deploy
func (this *DeployBuilder) setCMAnnotation(configStr string) {
	this.deploy.Spec.Template.Annotations[CMAnnotation] = configStr
}
func (d *DeployBuilder) apply() *DeployBuilder {
	*d.deploy.Spec.Replicas = int32(d.config.Spec.Replicas) //用户提交的配置yaml文件
	return d
}

// 所有者资源对象被删除时，与之相关的被拥有者资源对象也会被自动删除
func (d *DeployBuilder) setOwner() *DeployBuilder {
	//fmt.Printf("setOwner%+v\n", d.config)
	d.deploy.OwnerReferences = []v1.OwnerReference{
		{
			APIVersion: d.config.APIVersion,
			Kind:       d.config.Kind,
			Name:       d.config.Name,
			UID:        d.config.UID,
		},
	}
	return d
}

func (d *DeployBuilder) Replicas(r int) *DeployBuilder {
	*d.deploy.Spec.Replicas = int32(r)
	return d
}

func (d *DeployBuilder) Build(ctx context.Context) error {
	if d.deploy.CreationTimestamp.IsZero() {
		d.apply().setOwner()
		//先创建configmap
		err := d.cmBuilder.Build(ctx)
		if err != nil {
			return err
		}
		//设置 config md5
		d.setCMAnnotation(d.cmBuilder.DataKey)
		//后创建deployment
		err = d.Create(ctx, d.deploy)
		if err != nil {
			return err
		}
	} else {
		err := d.cmBuilder.Build(ctx) //更新configmap
		if err != nil {
			return err
		}
		//patch 也可以直接update
		patch := client.MergeFrom(d.deploy.DeepCopy()) //深度拷贝原来的deploy
		d.apply()                                      //同步  所需要的属性 如 副本数
		d.setCMAnnotation(d.cmBuilder.DataKey)         // 同步configmap
		err = d.Patch(ctx, d.deploy, patch)            //d.deploy是新的，patch是原来的
		//查看状态
		replicas := d.deploy.Status.ReadyReplicas
		d.config.Status.Ready = fmt.Sprintf("%d/%d", replicas, d.config.Spec.Replicas)
		d.config.Status.Replicas = replicas
		err = d.Client.Status().Update(ctx, d.config)
		if err != nil {
			return err
		}
	}
	return nil
}
