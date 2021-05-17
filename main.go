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
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"os"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/v6/controller"
)

var _ controller.Provisioner = &HostPathProvisioner{}

func main() {
	basePath := os.Getenv("PROVISIONER_PATH")
	if basePath == "" {
		basePath = "/persistentvolumes"
	}

	// get in cluster configuration
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("error setup in cluster config: %v", err)
	}

	// create k8s client instance
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("error create client: %v", err)
	}

	// get server version
	serverVersion, err := client.Discovery().ServerVersion()
	if err != nil {
		log.Fatalf("error getting server version: %v", err)
	}

	// configure controller implementation
	k8sProvisioner := &HostPathProvisioner{
		client:    client,
		localPath: basePath,
	}

	// initialize and run k8s controller instance
	k8sController := controller.NewProvisionController(
		client,
		ProvisionerName,
		k8sProvisioner,
		serverVersion.GitVersion,
		controller.LeaderElection(true),
	)
	k8sController.Run(context.Background())
}
