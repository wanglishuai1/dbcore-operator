package k8sconfig

import (
	"fmt"
	v1 "github.com/shenyisyn/dbcore/pkg/apis/dbconfig/v1"
	"github.com/shenyisyn/dbcore/pkg/controllers"
	"github.com/shenyisyn/dbcore/pkg/dashboard"
	appv1 "k8s.io/api/apps/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func InitManager() {
	logf.SetLogger(zap.New())
	mgr, err := manager.New(K8sRestConfig(), manager.Options{
		Logger: logf.Log.WithName("dbcore"),
	})
	if err != nil {
		mgr.GetLogger().Error(err, "unable to set up manager")
		os.Exit(1)
	}
	if err = v1.SchemeBuilder.AddToScheme(mgr.GetScheme()); err != nil {
		mgr.GetLogger().Error(err, "unable add scheme")
		os.Exit(1)
	}
	dbConfigController := controllers.NewDbConfigController(mgr.GetClient(), mgr.GetEventRecorderFor("dbconfig"))

	if err = builder.ControllerManagedBy(mgr).
		For(&v1.DbConfig{}).
		WatchesRawSource(
			source.Kind(mgr.GetCache(), &appv1.Deployment{}),
			handler.Funcs{
				DeleteFunc: dbConfigController.OnDelete,
				UpdateFunc: dbConfigController.OnUpdate,
			},
		).
		Complete(dbConfigController); err != nil {
		mgr.GetLogger().Error(err, "unable to create manager")
		os.Exit(1)
	}
	if err = mgr.Add(dashboard.NewAdminUi(mgr.GetClient())); err != nil {
		mgr.GetLogger().Error(err, "unable to add dashboard")
		os.Exit(1)
	}
	if err = mgr.Start(signals.SetupSignalHandler()); err != nil {
		mgr.GetLogger().Error(err, "unable to start manager")
		os.Exit(1)
	}
	fmt.Println(777)

}
