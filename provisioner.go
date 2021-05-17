/*
Copyright 2021 Christian Aye

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"os"
	"path/filepath"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/v6/controller"
	"strings"
)

const (
	ProvisionerName = "cluster.local/k8s-hostpath-provisioner"
)

func GeneratePVCName(pvcNamespace string, pvcName string, pvName string, namingStrategy string) string {
	if namingStrategy == "static" {
		return strings.Join([]string{pvcNamespace, pvcName}, "-")
	} else {
		return strings.Join([]string{pvcNamespace, pvcName, pvName}, "-")
	}
}

type HostPathProvisioner struct {
	client    kubernetes.Interface
	localPath string
}

func (p *HostPathProvisioner) Provision(_ context.Context, options controller.ProvisionOptions) (*v1.PersistentVolume, controller.ProvisioningState, error) {
	if options.PVC.Spec.Selector != nil {
		return nil, controller.ProvisioningFinished, fmt.Errorf("claim Selector is not supported")
	}

	log.Printf("new provision detected: VolumeOptions %+v", options)

	// build PV name
	pvcNamespace := options.PVC.Namespace
	pvcName := options.PVC.Name
	pvSubPath, exists := options.StorageClass.Parameters["subPath"]
	if !exists {
		return nil, controller.ProvisioningFinished, fmt.Errorf("subPath must be set in storage class")
	}

	namingStrategy := options.StorageClass.Parameters["namingStrategy"]
	pvName := GeneratePVCName(pvcNamespace, pvcName, string(options.PVC.UID), namingStrategy)

	log.Printf("create persistent volume: %s", pvName)

	// create directory for volume
	hostPath := filepath.Join(p.localPath, pvSubPath, pvName)

	log.Printf("create host directory (0777): %s", hostPath)

	if err := os.MkdirAll(hostPath, 0777); err != nil {
		return nil, controller.ProvisioningFinished, fmt.Errorf("unable to create directory for new pv: %v", err)
	}
	if err := os.Chmod(hostPath, 0777); err != nil {
		return nil, controller.ProvisioningFinished, fmt.Errorf("unable to change permission for new directory: %v", err)
	}

	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: pvName,
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: *options.StorageClass.ReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			MountOptions:                  options.StorageClass.MountOptions,
			Capacity: v1.ResourceList{
				v1.ResourceStorage: options.PVC.Spec.Resources.Requests[v1.ResourceStorage],
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: hostPath,
				},
			},
		},
	}

	log.Printf("new persistence volume: %v", pv)

	return pv, controller.ProvisioningFinished, nil
}

func (p *HostPathProvisioner) Delete(ctx context.Context, volume *v1.PersistentVolume) error {
	storageClass, err := p.client.StorageV1().StorageClasses().Get(ctx, volume.Spec.StorageClassName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to obtain storage class for volume %s", volume.Name)
	}

	hostPath := volume.Spec.HostPath.Path
	if _, err := os.Stat(hostPath); os.IsNotExist(err) {
		log.Printf("path %s does not exist, deletion skipped", hostPath)
		return nil
	}

	onDelete := storageClass.Parameters["onDelete"]
	if onDelete == "delete" {
		return os.RemoveAll(hostPath)
	} else if onDelete == "archive" {
		return os.Rename(hostPath, filepath.Join(p.localPath, volume.Name+"-archived"))
	}

	return nil
}
