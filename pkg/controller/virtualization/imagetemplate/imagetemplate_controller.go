/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com.tw
*/

package imagetemplate

import (
	"context"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	virtzv1alpha1 "kubesphere.io/api/virtualization/v1alpha1"
	"kubevirt.io/client-go/kubecli"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

const (
	controllerName        = "imagetemplate-controller"
	successSynced         = "Synced"
	messageResourceSynced = "ImageTemplate synced successfully"
)

// Reconciler reconciles a image template object
type Reconciler struct {
	client.Client
	Logger                  logr.Logger
	Recorder                record.EventRecorder
	MaxConcurrentReconciles int
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	if r.Logger.GetSink() == nil {
		r.Logger = ctrl.Log.WithName("controllers").WithName(controllerName)
	}
	if r.Recorder == nil {
		r.Recorder = mgr.GetEventRecorderFor(controllerName)
	}
	if r.MaxConcurrentReconciles <= 0 {
		r.MaxConcurrentReconciles = 1
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.MaxConcurrentReconciles,
		}).
		For(&virtzv1alpha1.ImageTemplate{}).
		Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.V(2).Infof("Reconciling ImageTemplate %s/%s", req.Namespace, req.Name)

	rootCtx := context.Background()
	imageTemplate := &virtzv1alpha1.ImageTemplate{}
	if err := r.Get(rootCtx, req.NamespacedName, imageTemplate); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	imageTemplate_instance := imageTemplate.DeepCopy()

	clientConfig := kubecli.DefaultClientConfig(&pflag.FlagSet{})

	// get the kubevirt client, using which kubevirt resources can be managed.
	virtClient, err := kubecli.GetKubevirtClientFromClientConfig(clientConfig)
	if err != nil {
		klog.Infof("Cannot obtain KubeVirt client: %v\n", err)
		return ctrl.Result{}, err
	}

	if !imageTemplate_instance.Status.Created {
		// Create data volume

		blockOwnerDeletion := true
		controller := true

		dv := &cdiv1.DataVolume{
			TypeMeta: metav1.TypeMeta{
				Kind: "DataVolume",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      imageTemplate.Name,
				Namespace: imageTemplate.Namespace,
			},
			Spec: cdiv1.DataVolumeSpec{
				PVC: &corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: imageTemplate.Spec.Resources.Requests,
					},
				},
			},
		}

		if imageTemplate.Spec.Source.HTTP != nil {
			dv.Spec.Source = &cdiv1.DataVolumeSource{
				HTTP: &cdiv1.DataVolumeSourceHTTP{
					URL: imageTemplate.Spec.Source.HTTP.URL,
				},
			}
		}

		if imageTemplate.Spec.Source.Clone != nil {
			dv.Spec.Source = &cdiv1.DataVolumeSource{
				PVC: &cdiv1.DataVolumeSourcePVC{
					Name:      imageTemplate.Spec.Source.Clone.Name,
					Namespace: imageTemplate.Spec.Source.Clone.Namespace,
				},
			}
		}

		dv.OwnerReferences = []metav1.OwnerReference{
			{
				APIVersion:         imageTemplate.APIVersion,
				BlockOwnerDeletion: &blockOwnerDeletion,
				Controller:         &controller,
				Kind:               imageTemplate.Kind,
				Name:               imageTemplate.Name,
				UID:                imageTemplate.UID,
			},
		}

		if _, err = virtClient.CdiClient().CdiV1beta1().DataVolumes(imageTemplate.Namespace).Create(rootCtx, dv, metav1.CreateOptions{}); err != nil {
			if errors.IsAlreadyExists(err) {
				klog.Infof("DataVolume %s/%s already exists", dv.Namespace, dv.Name)
			} else {
				klog.Infof("Cannot create DataVolume: %v\n", err)
				return ctrl.Result{}, err
			}
		}

		if dv.Status.Phase != cdiv1.Succeeded {
			klog.Infof("DataVolume %s/%s is not ready", dv.Namespace, dv.Name)
			imageTemplate_instance.Status.Ready = false
		}

		imageTemplate_instance.Status.Created = true
		klog.Infof("DataVolume %s/%s created", dv.Namespace, dv.Name)

		// Update status
		if err := r.Status().Update(rootCtx, imageTemplate_instance); err != nil {
			return ctrl.Result{}, err
		}

	}

	if !reflect.DeepEqual(imageTemplate, imageTemplate_instance) {
		if err := r.Update(rootCtx, imageTemplate_instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	// get data volume's status
	dv, err := virtClient.CdiClient().CdiV1beta1().DataVolumes(imageTemplate.Namespace).Get(rootCtx, imageTemplate.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		klog.Infof("Cannot get DataVolume: %v\n", err)
		return ctrl.Result{}, err
	}

	wait := time.Duration(0)

	if dv.Status.Phase != cdiv1.Succeeded {
		klog.Infof("DataVolume %s/%s is not ready, progress %s", dv.Namespace, dv.Name, dv.Status.Progress)
		imageTemplate_instance.Status.Ready = false
		wait = time.Duration(10) * time.Second
	} else {
		klog.Infof("DataVolume %s/%s is ready", dv.Namespace, dv.Name)
		imageTemplate_instance.Status.Ready = true
	}

	// Update status
	if err := r.Status().Update(rootCtx, imageTemplate_instance); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: wait}, nil
}
