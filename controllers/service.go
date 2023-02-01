package controllers

import (
	"context"
	v1 "github.com/tanlay/application-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *ApplicationReconciler) reconcileService(ctx context.Context, app *v1.Application) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	svc := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: app.Namespace,
		Name:      app.Name,
	}, svc)
	if err == nil {
		l.Info("The Service has already exist")
		if reflect.DeepEqual(svc.Status, app.Status.Network) {
			return ctrl.Result{}, nil
		}
		app.Status.Network = svc.Status
		if err := r.Status().Update(ctx, app); err != nil {
			l.Error(err, "Failed to update Application status")
			return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
		}
		l.Info("The Application status has benn updated")
		return ctrl.Result{}, nil
	}
	if !errors.IsNotFound(err) {
		l.Error(err, "Failed to get Service, will requeue after a short time")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}
	newSvc := &corev1.Service{}
	newSvc.SetName(app.Name)
	newSvc.SetNamespace(app.Namespace)
	newSvc.Spec = app.Spec.Service.ServiceSpec
	newSvc.Spec.Selector = app.Labels

	if err := ctrl.SetControllerReference(app, newSvc, r.Scheme); err != nil {
		l.Error(err, "Failed to SetControllerReference, will requeue after a short time")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}
	if err := r.Create(ctx, newSvc); err != nil {
		l.Error(err, "Failed to create Service, will requeue after a short time")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}
	l.Info("The Service has been created")
	return ctrl.Result{}, nil
}
