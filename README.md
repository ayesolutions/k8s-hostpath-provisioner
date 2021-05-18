# Kubernetes hostpath provisioner

k8s-hostpath-provisioner is a simple storage provisioner to manage dynamic provisioning in Kubernetes on local storages.

## Getting started

You can configure multiple storage classes with different configurations on one parent host path. Each configuration use his own sub-directory.

### Create a storage class

```
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: storage-default
  annotations:
    storageclass.kubernetes.io/is-default-class: 'true'
provisioner: cluster.local/k8s-hostpath-provisioner
parameters:
  namingStrategy: dynamic # or 'static' to use the PVC name without resource UUID 
  onDelete: delete # or 'archive' to rename PV folder on delete to $PVC-archive 
  subPath: default # sub folder name for this storage class. parent is defined in storage controller.
reclaimPolicy: Delete
allowVolumeExpansion: true
volumeBindingMode: Immediate
```

### Deploy provisioner controller to your cluster

```
kind: Deployment
apiVersion: apps/v1
metadata:
  name: k8s-hostpath-provisioner
  labels:
    app.kubernetes.io/name: k8s-hostpath-provisioner
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: k8s-hostpath-provisioner
  template:
    metadata:
      labels:
        app.kubernetes.io/name: k8s-hostpath-provisioner
    spec:
      restartPolicy: Always
      containers:
        - name: hostpath-provisioner
          image: 'ayesolutions/k8s-hostpath-provisioner'
          imagePullPolicy: Always
          env:
            - name: PROVISIONER_PATH
              value: /container/data/k8s
          resources: {}
          volumeMounts:
            - name: data
              mountPath: /container/data/k8s
      volumes:
        - name: data
          hostPath:
            path: /host/data/k8s
```
