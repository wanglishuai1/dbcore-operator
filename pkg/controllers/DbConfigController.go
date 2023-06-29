package controllers

import (
	"context"
	v1 "github.com/shenyisyn/dbcore/pkg/apis/dbconfig/v1"
	"github.com/shenyisyn/dbcore/pkg/builders"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	Kind            = "DbConfig"
	GroupApiVersion = "api.jtthink.com/v1"
)

type DbConfigController struct {
	client.Client
	E record.EventRecorder //记录事件
}

func NewDbConfigController(client client.Client, e record.EventRecorder) *DbConfigController {
	return &DbConfigController{
		Client: client,
		E:      e,
	}
}

func (r *DbConfigController) OnDelete(ctx context.Context, event event.DeleteEvent, limitingInterface workqueue.RateLimitingInterface) {
	for _, ref := range event.Object.GetOwnerReferences() {
		if ref.Kind == Kind && ref.APIVersion == GroupApiVersion {
			limitingInterface.Add(reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: event.Object.GetNamespace(),
					Name:      ref.Name,
				},
			})
		}
	}

}
func (r *DbConfigController) OnUpdate(ctx context.Context, event event.UpdateEvent, limitingInterface workqueue.RateLimitingInterface) {
	for _, ref := range event.ObjectNew.GetOwnerReferences() {
		if ref.Kind == Kind && ref.APIVersion == GroupApiVersion {
			limitingInterface.Add(reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: event.ObjectNew.GetNamespace(),
					Name:      ref.Name,
				},
			})
		}
	}

}
func (r *DbConfigController) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	config := &v1.DbConfig{}
	err := r.Get(ctx, req.NamespacedName, config)
	if err != nil {
		return reconcile.Result{}, err
	}
	builder, err := builders.NewDeployBuilder(*config, r.Client)
	r.E.Event(config, corev1.EventTypeNormal, "deployment", "开始构建")
	if err != nil {
		return reconcile.Result{}, err
	}
	err = builder.Build(ctx)

	if err != nil {
		r.E.Event(config, corev1.EventTypeWarning, "Build", err.Error())
		return reconcile.Result{}, err
	}
	r.E.Event(config, corev1.EventTypeNormal, "Build", "构建完成")
	return reconcile.Result{}, err
}
func (r *DbConfigController) InjectClient(c client.Client) error {
	r.Client = c
	return nil
}
