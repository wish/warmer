# warmer [![Docker Repository on Quay](https://quay.io/repository/wish/warmer/status "Docker Repository on Quay")](https://quay.io/repository/wish/warmer)

warmer is utility program that warms EBS drive that was restored from snapshot
by reading all files in order of their physical location on EBS to maximize
throughput on restore.

## Example usage
We recommend deploying `warmer` as `initContainer` to allow for transparent warming without need to modify your main image.

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata: 
  name: mongod
spec: 
  replicas: 1
  serviceName: mongod-svc
    spec: 
      containers: 
        image: mongo:3.4.15-jessie
        name: mongod
        volumeMounts: 
        - mountPath: /data
          name: data
      initContainers: 
      - command: ["/root/warmer", "/data"]
        image: quay.io/wish/warmer:latest
        name: warmer
        volumeMounts: 
        - mountPath: /data
          name: data
      volumes: 
      - name: data
        persistentVolumeClaim: 
          claimName: mongod-pvc

```
