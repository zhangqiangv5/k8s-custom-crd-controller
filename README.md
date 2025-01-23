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
### 在集群上运行
```azure
$ kubectl apply -f crd/network.yaml 
customresourcedefinition.apiextensions.k8s.io/networks.samplecrd.k8s.io created
$ kubectl get crd networks.samplecrd.k8s.io
NAME                        CREATED AT
networks.samplecrd.k8s.io   2025-01-23T14:51:59Z
```