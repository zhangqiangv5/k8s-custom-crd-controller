# k8s-custom-crd-controller
k8s-custom-crd-controller example

使用方法：
 ```
 go mod tidy # 安装依赖包
 go mod vendor # 安装包拷贝到vendor目录下
./hack/update-codegen.sh # 生成代码
```
执行上述命令后会在pkg/目录生成generated目录和在
pkg/apis/samplecrd/v1/zz_generated.deepcopy.go文件

修改crd后可以删除上面生成的文件和目录重新执行./hack/update-codegen.sh
### 在集群上运行自定义的crd
```azure
$ kubectl apply -f crd/network.yaml 
customresourcedefinition.apiextensions.k8s.io/networks.samplecrd.k8s.io created

$ kubectl get crd networks.samplecrd.k8s.io
NAME                        CREATED AT
networks.samplecrd.k8s.io   2025-01-23T14:51:59Z

$ kubectl apply -f example/example-network.yaml
network.samplecrd.k8s.io/example-network created

$ kubectl get network
    NAME              AGE
example-network   22s

# 通过 kubectl describe 命令，看到这个 Network 对象的细节
$ kubectl describe network example-network
Name:         example-network
Namespace:    default
Labels:       <none>
Annotations:  <none>
API Version:  samplecrd.k8s.io/v1
Kind:         Network
Metadata:
    Creation Timestamp:  2025-01-23T15:08:40Z
Generation:          1
  Resource Version:    206519
UID:                 ce18fdc5-b694-43b9-a899-b61ccb04e2cd
Spec:
Cidr:             192.16.0.0/16
  Deployment Name:  example-network
Gateway:          192.16.0.1
Replicas:         1
Events:             <none>
```