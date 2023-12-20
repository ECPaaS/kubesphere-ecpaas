/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com.tw
*/

package virtualmachine

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	snapv1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	kvapi "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/kubecli"

	virtzv1alpha1 "kubesphere.io/api/virtualization/v1alpha1"
)

const (
	controllerName                        = "virtualmachine-controller"
	successSynced                         = "Synced"
	messageResourceSynced                 = "VirtualMachine synced successfully"
	pvcCreateByDiskVolumeTemplatePrefix   = "tpl-" // tpl: template
	pvcCreateByDiskVolumeControllerPrefix = "new-"
	volumeSnapshotClassName               = "cstor-csi-disk"
)

// Reconciler reconciles a cnat object
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
	if r.Logger == nil {
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
		For(&virtzv1alpha1.VirtualMachine{}).
		Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.V(2).Infof("Reconciling VirtualMachine %s/%s", req.Namespace, req.Name)

	rootCtx := context.Background()

	vsc := snapv1.VolumeSnapshotClass{}
	if err := r.Get(rootCtx, client.ObjectKey{Name: volumeSnapshotClassName}, &vsc); err != nil {
		if errors.IsNotFound(err) {
			klog.Infof("VolumeSnapshotClass %s not found", volumeSnapshotClassName)
		} else {
			klog.Errorf("Failed to get VolumeSnapshotClass %s: %v", volumeSnapshotClassName, err)
			return ctrl.Result{}, err
		}
	}
	klog.V(2).Infof("VolumeSnapshotClass %s delete policy: %s", volumeSnapshotClassName, vsc.DeletionPolicy)

	if vsc.DeletionPolicy != "Retain" {
		klog.Infof("VolumeSnapshotClass %s delete policy is not Retain", volumeSnapshotClassName)
		vsc_instance := vsc.DeepCopy()
		vsc_instance.DeletionPolicy = "Retain"
		if err := r.Update(rootCtx, vsc_instance); err != nil {
			klog.Errorf("Failed to update VolumeSnapshotClass %s : %v", volumeSnapshotClassName, err)
			return ctrl.Result{}, err
		}
		klog.Infof("VolumeSnapshotClass %s delete policy is updated to Retain", volumeSnapshotClassName)
	}

	vm := &virtzv1alpha1.VirtualMachine{}
	if err := r.Get(rootCtx, req.NamespacedName, vm); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	vm_instance := vm.DeepCopy()

	clientConfig := kubecli.DefaultClientConfig(&pflag.FlagSet{})

	// get the kubevirt client, using which kubevirt resources can be managed.
	virtClient, err := kubecli.GetKubevirtClientFromClientConfig(clientConfig)
	if err != nil {
		klog.Infof("Cannot obtain KubeVirt client: %v\n", err)
		return ctrl.Result{}, err
	}

	if IsDeletionCandidate(vm_instance, virtzv1alpha1.VirtualMachineFinalizer) {
		klog.Infof("Deleting VirtualMachine %s/%s", req.Namespace, req.Name)

		if err := deleteVirtualMachine(virtClient, req.Namespace, vm_instance); err != nil {
			return ctrl.Result{}, err
		}

		err = r.deleteDiskVolumeOwnerLabelInVMDiskVolumes(vm_instance, req)
		if err != nil {
			return ctrl.Result{}, err
		}

		klog.Infof("Removing finalizer for VirtualMachine %s/%s", req.Namespace, req.Name)
		if err := r.removeFinalizer(vm_instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	if NeedToAddFinalizer(vm_instance, virtzv1alpha1.VirtualMachineFinalizer) || !vm_instance.Status.Created {
		klog.Infof("Adding finalizer for VirtualMachine %s/%s", req.Namespace, req.Name)
		if err := r.addFinalizer(vm_instance); err != nil {
			return ctrl.Result{}, err
		}

		// create disk volume
		if vm_instance.Spec.DiskVolumeTemplates != nil {
			klog.Infof("Creating DiskVolume for VirtualMachine %s/%s", req.Namespace, req.Name)

			for _, diskVolumeTemplate := range vm_instance.Spec.DiskVolumeTemplates {
				diskVolume := GenerateDiskVolume(vm_instance, &diskVolumeTemplate)

				err := r.Create(rootCtx, diskVolume)
				if err != nil {
					if reflect.TypeOf(err) == reflect.TypeOf(&errors.StatusError{}) {
						statusErr := err.(*errors.StatusError)
						if statusErr.ErrStatus.Reason == metav1.StatusReasonAlreadyExists {
							klog.Infof("DiskVolume %s/%s already exists", req.Namespace, diskVolume.Name)
						}
					} else {
						klog.Infof(err.Error())
						return ctrl.Result{}, err
					}
				}
			}
		}

		// add disk volume owner label in order to update disk volume status
		if vm_instance.Spec.DiskVolumes != nil {
			for _, diskVolume := range vm_instance.Spec.DiskVolumes {
				dv := &virtzv1alpha1.DiskVolume{}
				if err := r.Get(rootCtx, types.NamespacedName{Name: diskVolume, Namespace: req.Namespace}, dv); err != nil {
					return ctrl.Result{}, err
				}

				if dv.Labels[virtzv1alpha1.VirtualizationDiskType] == "data" {
					err := r.addDiskVolumeOwnerLabel(diskVolume, req.Namespace, vm_instance.Name)
					if err != nil {
						return ctrl.Result{}, err
					}
				}
			}
		}

		klog.Infof("Creating VirtualMachine %s/%s", req.Namespace, req.Name)
		err := createVirtualMachine(virtClient, vm_instance)

		if err != nil && reflect.TypeOf(err) == reflect.TypeOf(&errors.StatusError{}) {
			statusErr := err.(*errors.StatusError)
			if statusErr.ErrStatus.Reason == metav1.StatusReasonAlreadyExists {
				klog.Infof("VirtualMachine %s/%s already exists", req.Namespace, vm_instance.Name)
			} else {
				klog.Infof(err.Error())
				return ctrl.Result{}, err
			}
		}
	}

	// add or remove disk volume to kubevirt VM, based on spec.diskvolumes
	err = r.updateDiskVolumes(vm_instance, req, virtClient)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := getVirtualMachineStatus(virtClient, req.Namespace, vm_instance); err != nil {
		return ctrl.Result{}, err
	}

	// update status, refresh the status even when the virtualmachine is not ready
	if !reflect.DeepEqual(vm.Status, vm_instance.Status) {
		if err := r.Status().Update(rootCtx, vm_instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	if strings.Title(vm_instance.Spec.RunStrategy) == string(kvapi.RunStrategyAlways) {
		if vm_instance.Status.PrintableStatus != kvapi.VirtualMachineStatusRunning {
			klog.V(2).Infof("VirtualMachine %s/%s is not running", req.Namespace, req.Name)
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
		if !vm_instance.Status.Ready {
			klog.V(2).Infof("VirtualMachine %s/%s is not ready", req.Namespace, req.Name)
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
	} else if strings.Title(vm_instance.Spec.RunStrategy) == string(kvapi.RunStrategyHalted) {
		if vm_instance.Status.PrintableStatus != kvapi.VirtualMachineStatusStopped {
			klog.V(2).Infof("VirtualMachine %s/%s is not stopped", req.Namespace, req.Name)
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
	} else {
		klog.V(2).Infof("VirtualMachine %s/%s runStrategy is not supported", req.Namespace, req.Name)
		return ctrl.Result{}, nil
	}

	// delete volumesnapshotcontent
	if vm_instance.Spec.DiskVolumeTemplates != nil {
		for _, diskVolumeTemplate := range vm_instance.Spec.DiskVolumeTemplates {
			volumesnapshotcontents := &snapv1.VolumeSnapshotContentList{}
			if err := r.List(rootCtx, volumesnapshotcontents); err != nil {
				return ctrl.Result{}, err
			}

			for _, volumesnapshotcontent := range volumesnapshotcontents.Items {
				if strings.HasPrefix(volumesnapshotcontent.Spec.VolumeSnapshotRef.Name, pvcCreateByDiskVolumeTemplatePrefix+diskVolumeTemplate.Name) {
					klog.Infof("Deleting VolumeSnapshotContent %s", volumesnapshotcontent.Name)
					if err := r.Delete(rootCtx, &volumesnapshotcontent); err != nil {
						return ctrl.Result{}, err
					}
				}
			}
		}
	}

	// update last disk volumes
	vm_instance.Annotations[virtzv1alpha1.VirtualizationLastDiskVolumes] = strings.Join(vm_instance.Spec.DiskVolumes, ",")

	if !reflect.DeepEqual(vm, vm_instance) {
		if err := r.Update(rootCtx, vm_instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	// update status
	if !reflect.DeepEqual(vm.Status, vm_instance.Status) {
		if err := r.Status().Update(rootCtx, vm_instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	// update event
	r.Recorder.Event(vm, corev1.EventTypeNormal, successSynced, messageResourceSynced)
	return ctrl.Result{}, nil

}

func (r *Reconciler) updateDiskVolumes(vm_instance *virtzv1alpha1.VirtualMachine, req ctrl.Request, virtClient kubecli.KubevirtClient) error {
	// get all data disk volume name from disk volume template
	diskVolumeTemplateDataVolumeNames := make([]string, len(vm_instance.Spec.DiskVolumeTemplates))
	for i, diskVolumeTemplate := range vm_instance.Spec.DiskVolumeTemplates {
		if diskVolumeTemplate.Labels[virtzv1alpha1.VirtualizationDiskType] == "data" {
			diskVolumeTemplateDataVolumeNames[i] = diskVolumeTemplate.Name
		}
	}

	if vm_instance.Spec.DiskVolumes != nil {
		if vm_instance.Annotations[virtzv1alpha1.VirtualizationLastDiskVolumes] != "" {
			lastDiskVolumes := strings.Split(vm_instance.Annotations[virtzv1alpha1.VirtualizationLastDiskVolumes], ",")
			for _, lastDiskVolume := range lastDiskVolumes {
				// skip system disk
				if vm_instance.Annotations[virtzv1alpha1.VirtualizationSystemDiskName] == lastDiskVolume {
					continue
				}

				if !ContainsString(vm_instance.Spec.DiskVolumes, lastDiskVolume, nil) {
					klog.Infof("Removing DiskVolume %s/%s from VirtualMachine %s/%s", req.Namespace, lastDiskVolume, req.Namespace, req.Name)
					err := removeVolume(vm_instance.Name, lastDiskVolume, vm_instance.Namespace, virtClient)
					if err != nil {
						klog.V(2).Infof(err.Error())
						return err
					}
					err = r.removeDiskVolumeOwnerLabel(lastDiskVolume, req.Namespace)
					if err != nil {
						return err
					}
				}
			}
		}

		for _, diskVolume := range vm_instance.Spec.DiskVolumes {
			// skip system disk
			if vm_instance.Annotations[virtzv1alpha1.VirtualizationSystemDiskName] == diskVolume {
				continue
			}

			// skip hotpluggable is false
			dv := &virtzv1alpha1.DiskVolume{}
			if err := r.Get(context.Background(), types.NamespacedName{Name: diskVolume, Namespace: req.Namespace}, dv); err != nil {
				return err
			}
			if dv.Labels[virtzv1alpha1.VirtualizationDiskHotpluggable] == "false" {
				continue
			}

			if vm_instance.Annotations[virtzv1alpha1.VirtualizationLastDiskVolumes] != "" {
				lastDiskVolumes := strings.Split(vm_instance.Annotations[virtzv1alpha1.VirtualizationLastDiskVolumes], ",")
				if ContainsString(lastDiskVolumes, diskVolume, nil) {
					continue
				}
			}

			klog.Infof("Adding DiskVolume %s/%s to VirtualMachine %s/%s", req.Namespace, diskVolume, req.Namespace, req.Name)
			err := addVolume(vm_instance.Name, diskVolume, vm_instance.Namespace, virtClient)
			if err != nil {
				klog.V(2).Infof(err.Error())

				if reflect.TypeOf(err) == reflect.TypeOf(&errors.StatusError{}) {
					statusErr := err.(*errors.StatusError)
					if statusErr.ErrStatus.Reason == metav1.StatusReasonAlreadyExists {
						klog.Infof("DiskVolume %s/%s already exists", req.Namespace, diskVolume)
					} else {
						klog.V(2).Infof(err.Error())
						return err
					}
				}
			}
			err = r.addDiskVolumeOwnerLabel(diskVolume, req.Namespace, vm_instance.Name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Reconciler) deleteDiskVolumeOwnerLabelInVMDiskVolumes(vm_instance *virtzv1alpha1.VirtualMachine, req ctrl.Request) error {

	if vm_instance.Spec.DiskVolumes != nil {
		for _, diskVolume := range vm_instance.Spec.DiskVolumes {
			err := r.removeDiskVolumeOwnerLabel(diskVolume, req.Namespace)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Reconciler) addDiskVolumeOwnerLabel(diskVolumeName, namespace string, vmName string) error {
	rootCtx := context.Background()

	dv := &virtzv1alpha1.DiskVolume{}
	if err := r.Get(rootCtx, types.NamespacedName{Name: diskVolumeName, Namespace: namespace}, dv); err != nil {
		return err
	}

	klog.V(2).Infof("Add DiskVolume %s/%s owner label", namespace, diskVolumeName)

	copy_dv := dv.DeepCopy()
	copy_dv.Labels[virtzv1alpha1.VirtualizationDiskVolumeOwner] = vmName

	if err := r.Update(rootCtx, copy_dv); err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) removeDiskVolumeOwnerLabel(diskVolumeName, namespace string) error {
	rootCtx := context.Background()

	dv := &virtzv1alpha1.DiskVolume{}
	if err := r.Get(rootCtx, types.NamespacedName{Name: diskVolumeName, Namespace: namespace}, dv); err != nil {
		return err
	}

	klog.V(2).Infof("Delete DiskVolume %s/%s owner label", namespace, diskVolumeName)

	copy_dv := dv.DeepCopy()
	delete(copy_dv.Labels, virtzv1alpha1.VirtualizationDiskVolumeOwner)

	if err := r.Update(rootCtx, copy_dv); err != nil {
		return err
	}

	return nil
}

func GenerateDiskVolume(vm_instance *virtzv1alpha1.VirtualMachine, diskVolumeTemplate *virtzv1alpha1.DiskVolume) *virtzv1alpha1.DiskVolume {

	blockOwnerDeletion := true
	controller := true

	diskVolume := &virtzv1alpha1.DiskVolume{}
	diskVolume.Name = diskVolumeTemplate.Name
	diskVolume.Namespace = diskVolumeTemplate.Namespace
	diskVolume.Annotations = diskVolumeTemplate.Annotations
	diskVolume.Labels = diskVolumeTemplate.Labels
	diskVolume.Spec.PVCName = pvcCreateByDiskVolumeTemplatePrefix + diskVolumeTemplate.Name
	diskVolume.Spec.Resources = diskVolumeTemplate.Spec.Resources
	diskVolume.Spec.Source = diskVolumeTemplate.Spec.Source

	// For check data volume status
	if diskVolume.Annotations == nil {
		diskVolume.Annotations = make(map[string]string)
	}
	diskVolume.Annotations["cdi.kubevirt.io/storage.deleteAfterCompletion"] = "false"

	if diskVolumeTemplate.Labels[virtzv1alpha1.VirtualizationDiskType] == "system" ||
		diskVolumeTemplate.Labels[virtzv1alpha1.VirtualizationDiskMediaType] == "cdrom" {
		diskVolume.OwnerReferences = []metav1.OwnerReference{
			{
				APIVersion:         vm_instance.APIVersion,
				Kind:               vm_instance.Kind,
				Name:               vm_instance.Name,
				UID:                vm_instance.UID,
				Controller:         &controller,
				BlockOwnerDeletion: &blockOwnerDeletion,
			},
		}
	}

	return diskVolume
}

func (c *Reconciler) addFinalizer(virtualmachine *virtzv1alpha1.VirtualMachine) error {
	clone := virtualmachine.DeepCopy()
	controllerutil.AddFinalizer(clone, virtzv1alpha1.VirtualMachineFinalizer)

	err := c.Update(context.Background(), clone)
	if err != nil {
		klog.V(3).Infof("Error adding  finalizer to virtualmachine %s: %v", virtualmachine.Name, err)
		return err
	}
	klog.V(3).Infof("Added finalizer to virtualmachine %s", virtualmachine.Name)
	return nil
}

func (c *Reconciler) removeFinalizer(virtualmachine *virtzv1alpha1.VirtualMachine) error {
	clone := virtualmachine.DeepCopy()
	controllerutil.RemoveFinalizer(clone, virtzv1alpha1.VirtualMachineFinalizer)
	err := c.Update(context.Background(), clone)
	if err != nil {
		klog.V(3).Infof("Error removing  finalizer from virtualmachine %s: %v", virtualmachine.Name, err)
		return err
	}
	klog.V(3).Infof("Removed protection finalizer from virtualmachine %s", virtualmachine.Name)
	return nil
}

func getVirtualMachineStatus(virtClient kubecli.KubevirtClient, namespace string, vm *virtzv1alpha1.VirtualMachine) error {
	kv_vm := kvapi.VirtualMachine{}

	// get kubevirt virtualmachine
	err := virtClient.RestClient().Get().Namespace(namespace).Resource("virtualmachines").Name(vm.ObjectMeta.Name).Do(context.Background()).Into(&kv_vm)
	if err != nil && !errors.IsNotFound(err) {
		klog.V(3).Infof("Error getting virtualmachine: %v", err)
		return err
	}

	// get the virtualmachine status
	vm.Status = kv_vm.Status

	return nil

}

func applyVirtualMachineSpec(kvvmSpec *kvapi.VirtualMachineSpec, virtzSpec virtzv1alpha1.VirtualMachineSpec) {

	runStrategy := kvapi.RunStrategyAlways
	if virtzSpec.RunStrategy == virtzv1alpha1.VirtualMachineRunStrategyAlways {
		runStrategy = kvapi.RunStrategyAlways
	} else if virtzSpec.RunStrategy == virtzv1alpha1.VirtualMachineRunStrategyHalted {
		runStrategy = kvapi.RunStrategyHalted
	} else {
		klog.Infof("RunStrategy %s is not supported", virtzSpec.RunStrategy)
	}
	kvvmSpec.RunStrategy = &runStrategy

	kvvmSpec.Template = &kvapi.VirtualMachineInstanceTemplateSpec{}
	kvvmSpec.Template.Spec = kvapi.VirtualMachineInstanceSpec{}
	kvvmSpec.Template.Spec.Domain = kvapi.DomainSpec{}
	kvvmSpec.Template.Spec.Domain.Resources = kvapi.ResourceRequirements{}

	if virtzSpec.Hardware.Domain.Devices.Interfaces != nil {
		kvvmSpec.Template.Spec.Domain.Devices.Interfaces = make([]kvapi.Interface, len(virtzSpec.Hardware.Domain.Devices.Interfaces))
		for i, iface := range virtzSpec.Hardware.Domain.Devices.Interfaces {
			interfaceMehod := getInterfaceMethod(iface)
			kvvmSpec.Template.Spec.Domain.Devices.Interfaces[i] = kvapi.Interface{
				Name:                   iface.Name,
				InterfaceBindingMethod: interfaceMehod,
			}
		}
	}

	kvvmSpec.Template.Spec.Domain.Resources.Requests = virtzSpec.Hardware.Domain.Resources.Requests
	kvvmSpec.Template.Spec.Domain.CPU = &kvapi.CPU{}
	kvvmSpec.Template.Spec.Domain.CPU.Cores = virtzSpec.Hardware.Domain.CPU.Cores

	kvvmSpec.Template.Spec.Hostname = virtzSpec.Hardware.Hostname

	if virtzSpec.Hardware.Volumes != nil {
		kvvmSpec.Template.Spec.Volumes = make([]kvapi.Volume, len(virtzSpec.Hardware.Volumes))
		for i, volume := range virtzSpec.Hardware.Volumes {
			newDisk := kvapi.Disk{}
			newDisk.Name = volume.Name
			newDisk.DiskDevice = kvapi.DiskDevice{}
			newDisk.DiskDevice.Disk = &kvapi.DiskTarget{}
			newDisk.DiskDevice.Disk.Bus = "virtio"

			if volume.PersistentVolumeClaim != nil {
				kvvmSpec.Template.Spec.Volumes[i] = kvapi.Volume{
					Name: volume.Name,
					VolumeSource: kvapi.VolumeSource{
						PersistentVolumeClaim: &kvapi.PersistentVolumeClaimVolumeSource{
							PersistentVolumeClaimVolumeSource: corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: volume.PersistentVolumeClaim.ClaimName,
							},
						},
					},
				}

				kvvmSpec.Template.Spec.Domain.Devices.Disks = append(kvvmSpec.Template.Spec.Domain.Devices.Disks, newDisk)
			}
			if volume.CloudInitNoCloud != nil {
				kvvmSpec.Template.Spec.Volumes[i] = kvapi.Volume{
					Name: volume.Name,
					VolumeSource: kvapi.VolumeSource{
						CloudInitNoCloud: &kvapi.CloudInitNoCloudSource{
							UserDataBase64: volume.CloudInitNoCloud.UserDataBase64,
						},
					},
				}

				kvvmSpec.Template.Spec.Domain.Devices.Disks = append(kvvmSpec.Template.Spec.Domain.Devices.Disks, newDisk)
			}
			if volume.ContainerDisk != nil {
				kvvmSpec.Template.Spec.Volumes[i] = kvapi.Volume{
					Name: volume.Name,
					VolumeSource: kvapi.VolumeSource{
						ContainerDisk: &kvapi.ContainerDiskSource{
							Image: volume.ContainerDisk.Image,
						},
					},
				}

				kvvmSpec.Template.Spec.Domain.Devices.Disks = append(kvvmSpec.Template.Spec.Domain.Devices.Disks, newDisk)
			}
		}
	}

	if virtzSpec.Hardware.Domain.Machine != nil {
		kvvmSpec.Template.Spec.Domain.Machine = &kvapi.Machine{}
		kvvmSpec.Template.Spec.Domain.Machine = virtzSpec.Hardware.Domain.Machine
	}

	if virtzSpec.Hardware.Domain.Features != nil {
		kvvmSpec.Template.Spec.Domain.Features = &kvapi.Features{}
		kvvmSpec.Template.Spec.Domain.Features = virtzSpec.Hardware.Domain.Features
	}

	if virtzSpec.Hardware.Domain.Devices.Disks != nil {
		for _, disk := range virtzSpec.Hardware.Domain.Devices.Disks {
			newDisk := kvapi.Disk{}
			newDisk.Name = disk.Name
			newDisk.DiskDevice = kvapi.DiskDevice{}
			if disk.DiskDevice.Disk != nil {
				newDisk.DiskDevice.Disk = &kvapi.DiskTarget{}
				newDisk.DiskDevice.Disk = disk.DiskDevice.Disk.DeepCopy()
			} else if disk.DiskDevice.CDRom != nil {
				newDisk.DiskDevice.CDRom = &kvapi.CDRomTarget{}
				newDisk.DiskDevice.CDRom = disk.DiskDevice.CDRom.DeepCopy()
			}
			if disk.BootOrder != nil {
				boorOrder := *disk.BootOrder
				newDisk.BootOrder = &boorOrder
			}

			match := false
			for i, kvvm_disk := range kvvmSpec.Template.Spec.Domain.Devices.Disks {
				if kvvm_disk.Name == disk.Name {
					// replace disk
					kvvmSpec.Template.Spec.Domain.Devices.Disks[i] = newDisk
					match = true
					break
				}
			}

			if !match {
				kvvmSpec.Template.Spec.Domain.Devices.Disks = append(kvvmSpec.Template.Spec.Domain.Devices.Disks, newDisk)
			}
		}
	}

	if virtzSpec.Hardware.Domain.Devices.Inputs != nil {
		kvvmSpec.Template.Spec.Domain.Devices.Inputs = make([]kvapi.Input, len(virtzSpec.Hardware.Domain.Devices.Inputs))
		for i, input := range virtzSpec.Hardware.Domain.Devices.Inputs {
			kvvmSpec.Template.Spec.Domain.Devices.Inputs[i] = kvapi.Input{
				Type: input.Type,
				Bus:  input.Bus,
				Name: input.Name,
			}
		}
	}

	if virtzSpec.Hardware.Networks != nil {
		kvvmSpec.Template.Spec.Networks = make([]kvapi.Network, len(virtzSpec.Hardware.Networks))
		for i, network := range virtzSpec.Hardware.Networks {
			networkSource := getNetwork(network)
			kvvmSpec.Template.Spec.Networks[i] = kvapi.Network{
				Name:          network.Name,
				NetworkSource: networkSource,
			}
		}
	}

	if virtzSpec.DiskVolumes != nil {
		for _, volume := range virtzSpec.DiskVolumes {
			// check boot order from spec.diskvolumeTemplates label
			bootorder := uint(0)
			isMappingTodiskVolumeTemplate := false
			diskMediaType := "disk"
			isHotpluggable := false

			for _, diskVolumeTemplate := range virtzSpec.DiskVolumeTemplates {
				if diskVolumeTemplate.Name == volume {
					isMappingTodiskVolumeTemplate = true

					if diskVolumeTemplate.Labels != nil {
						val, ok := diskVolumeTemplate.Labels[virtzv1alpha1.VirtualizationBootOrder]
						if ok {
							uint64, _ := strconv.ParseUint(val, 10, 32)
							bootorder = uint(uint64)
						}
						val, ok = diskVolumeTemplate.Labels[virtzv1alpha1.VirtualizationDiskMediaType]
						if ok {
							diskMediaType = val
						}
						val, ok = diskVolumeTemplate.Labels[virtzv1alpha1.VirtualizationDiskHotpluggable]
						if ok {
							isHotpluggable, _ = strconv.ParseBool(val)
						}
					}
				}
			}

			if isMappingTodiskVolumeTemplate {
				newVolume := kvapi.Volume{
					Name: volume,
					VolumeSource: kvapi.VolumeSource{
						PersistentVolumeClaim: &kvapi.PersistentVolumeClaimVolumeSource{
							PersistentVolumeClaimVolumeSource: corev1.PersistentVolumeClaimVolumeSource{},
						},
					},
				}

				// system disk and cdrom disk is not hotpluggable
				if !isHotpluggable {
					newVolume.VolumeSource.PersistentVolumeClaim.ClaimName = pvcCreateByDiskVolumeTemplatePrefix + volume
					kvvmSpec.Template.Spec.Volumes = append(kvvmSpec.Template.Spec.Volumes, newVolume)

					newDisk := kvapi.Disk{}
					newDisk.Name = volume
					newDisk.BootOrder = &bootorder

					if diskMediaType == "cdrom" {
						newDisk.DiskDevice = kvapi.DiskDevice{
							CDRom: &kvapi.CDRomTarget{
								Bus: "sata",
							},
						}
					} else {
						newDisk.DiskDevice = kvapi.DiskDevice{
							Disk: &kvapi.DiskTarget{
								Bus: "virtio",
							},
						}
					}

					kvvmSpec.Template.Spec.Domain.Devices.Disks = append(kvvmSpec.Template.Spec.Domain.Devices.Disks, newDisk)
				}
			}
		}
	}

}

func getInterfaceMethod(iface kvapi.Interface) kvapi.InterfaceBindingMethod {
	interfaceMethod := kvapi.InterfaceBindingMethod{}

	klog.V(2).Infof("Interface %s", iface.Name)
	if iface.Bridge != nil {
		interfaceMethod.Bridge = &kvapi.InterfaceBridge{}
	} else if iface.Macvtap != nil {
		interfaceMethod.Macvtap = &kvapi.InterfaceMacvtap{}
	} else if iface.Masquerade != nil {
		interfaceMethod.Masquerade = &kvapi.InterfaceMasquerade{}
	} else if iface.SRIOV != nil {
		interfaceMethod.SRIOV = &kvapi.InterfaceSRIOV{}
	} else if iface.Slirp != nil {
		interfaceMethod.Slirp = &kvapi.InterfaceSlirp{}
	} else {
		// default assign interface to pod network.
		interfaceMethod.Masquerade = &kvapi.InterfaceMasquerade{}
	}

	return interfaceMethod
}

func getNetwork(network virtzv1alpha1.Network) kvapi.NetworkSource {
	networkSource := kvapi.NetworkSource{}

	if network.Pod != nil {
		networkSource.Pod = &kvapi.PodNetwork{
			VMNetworkCIDR:     network.Pod.VMNetworkCIDR,
			VMIPv6NetworkCIDR: network.Pod.VMIPv6NetworkCIDR,
		}
	} else if network.Multus != nil {
		networkSource.Multus = &kvapi.MultusNetwork{
			NetworkName: network.Multus.NetworkName,
			Default:     network.Multus.Default,
		}
	} else {
		// default assign interface to pod network.
		networkSource.Pod = &kvapi.PodNetwork{}
	}

	return networkSource
}

func createVirtualMachine(virtClient kubecli.KubevirtClient, virtzVM *virtzv1alpha1.VirtualMachine) error {

	blockOwnerDeletion := true
	controller := true

	namespace := "default"
	if virtzVM.Namespace != "" {
		namespace = virtzVM.Namespace
	}

	kvVM := &kvapi.VirtualMachine{
		TypeMeta: metav1.TypeMeta{
			Kind: "VirtualMachine",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      virtzVM.Name,
			Namespace: namespace,
		},
		Spec: kvapi.VirtualMachineSpec{},
	}
	kvVM.OwnerReferences = append(kvVM.OwnerReferences, metav1.OwnerReference{
		APIVersion:         virtzVM.APIVersion,
		BlockOwnerDeletion: &blockOwnerDeletion,
		Controller:         &controller,
		Kind:               virtzVM.Kind,
		Name:               virtzVM.Name,
		UID:                virtzVM.UID,
	})

	applyVirtualMachineSpec(&kvVM.Spec, virtzVM.Spec)

	createdVM, err := virtClient.VirtualMachine(namespace).Create(kvVM)
	if err != nil {
		klog.Infof(err.Error())
		return err
	}

	createdVM, err = virtClient.VirtualMachine(createdVM.Namespace).Get(createdVM.Name, &metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		klog.Infof(err.Error())
		return err
	}

	virtzVM.Status = createdVM.Status

	return nil
}

func deleteVirtualMachine(virtClient kubecli.KubevirtClient, namespace string, vm_instance *virtzv1alpha1.VirtualMachine) error {
	err := virtClient.VirtualMachine(namespace).Delete(vm_instance.Name, &metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		klog.Infof(err.Error())
		return err
	}
	return nil
}

// IsDeletionCandidate checks if object is candidate to be deleted
func IsDeletionCandidate(obj metav1.Object, finalizer string) bool {
	return obj.GetDeletionTimestamp() != nil && ContainsString(obj.GetFinalizers(),
		finalizer, nil)
}

// NeedToAddFinalizer checks if need to add finalizer to object
func NeedToAddFinalizer(obj metav1.Object, finalizer string) bool {
	return obj.GetDeletionTimestamp() == nil && !ContainsString(obj.GetFinalizers(),
		finalizer, nil)
}

// ContainsString checks if a given slice of strings contains the provided string.
// If a modifier func is provided, it is called with the slice item before the comparation.
func ContainsString(slice []string, s string, modifier func(s string) string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
		if modifier != nil && modifier(item) == s {
			return true
		}
	}
	return false
}

func addVolume(vmiName, volumeName, namespace string, virtClient kubecli.KubevirtClient) error {
	volumeSource, err := getVolumeSourceFromVolume(volumeName, namespace, virtClient)
	if err != nil {
		return fmt.Errorf("error adding volume, %v", err)
	}
	hotplugRequest := &kvapi.AddVolumeOptions{
		Name: volumeName,
		Disk: &kvapi.Disk{
			DiskDevice: kvapi.DiskDevice{
				Disk: &kvapi.DiskTarget{
					Bus: "scsi",
				},
			},
		},
		VolumeSource: volumeSource,
	}

	err = virtClient.VirtualMachine(namespace).AddVolume(vmiName, hotplugRequest)
	if err != nil {
		return fmt.Errorf("error adding volume, %v", err)
	}
	klog.Infof("Successfully submitted add volume request to VM %s for volume %s\n", vmiName, volumeName)
	return nil
}

func getVolumeSourceFromVolume(volumeName, namespace string, virtClient kubecli.KubevirtClient) (*kvapi.HotplugVolumeSource, error) {
	//Check if data volume exists.
	// _, err := virtClient.CdiClient().CdiV1beta1().DataVolumes(namespace).Get(context.TODO(), volumeName, metav1.GetOptions{})
	// if err == nil {
	// 	return &v1.HotplugVolumeSource{
	// 		DataVolume: &v1.DataVolumeSource{
	// 			Name:         volumeName,
	// 			Hotpluggable: true,
	// 		},
	// 	}, nil
	// }
	// DataVolume not found, try PVC

	// list all pvc
	pvcs, err := virtClient.CoreV1().PersistentVolumeClaims(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting pvc list, %v", err)
	}

	// find the pvc name contains volumeName
	targetPVCName := ""
	for _, pvc := range pvcs.Items {
		if strings.Contains(pvc.Name, volumeName) {
			targetPVCName = pvc.Name
			break
		}
	}

	_, err = virtClient.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), targetPVCName, metav1.GetOptions{})
	if err == nil {
		return &kvapi.HotplugVolumeSource{
			PersistentVolumeClaim: &kvapi.PersistentVolumeClaimVolumeSource{
				PersistentVolumeClaimVolumeSource: corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: targetPVCName,
				},
				Hotpluggable: true,
			},
		}, nil
	}

	if errors.IsNotFound(err) {
		return nil, fmt.Errorf("target pvc %s for volume %s not found, wait for created", targetPVCName, volumeName)
	}

	// Neither return error
	return nil, fmt.Errorf("volume %s is not a data volume or persistent volume claim", volumeName)
}

func removeVolume(vmiName, volumeName, namespace string, virtClient kubecli.KubevirtClient) error {
	err := virtClient.VirtualMachine(namespace).RemoveVolume(vmiName, &kvapi.RemoveVolumeOptions{
		Name: volumeName,
	})
	if err != nil {
		return fmt.Errorf("error removing volume, %v", err)
	}
	klog.Infof("Successfully submitted remove volume request to VM %s for volume %s\n", vmiName, volumeName)
	return nil
}
