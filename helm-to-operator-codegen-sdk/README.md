# helm-to-operator-codegen-sdk 
### Prerequisties:
1. GoLang Version: 1.21
```
wget -c https://golang.org/dl/go1.21.6.linux-amd64.tar.gz
sudo tar -C /usr/local -xvzf go1.21.6.linux-amd64.tar.gz
# Create Folder for Go Packages:
mkdir go
cd go
mkdir pkg bin src

# In order to add Go to environment: Append the following to "/etc/profile" file
export  PATH=$PATH:/usr/local/go/bin
export GOPATH="$HOME/go"
export GOBIN="$GOPATH/bin"

# Then, Apply the changes:
source /etc/profile
# Check The Installation
go version
```


2. Helm :
```
wget https://get.helm.sh/helm-v3.9.3-linux-amd64.tar.gz
tar xvf helm-v3.9.3-linux-amd64.tar.gz
sudo mv linux-amd64/helm /usr/local/bin
```

3. Go Packages:
```
# Clone the Repo
cd nephio-sdk/
go mod tidy
```

### Running the Script
```
go run main.go <path_to_local_helm_chart> <namespace>
e.g. go run main.go /home/ubuntu/free5gccharts/towards5gs-helm/charts/free5gc/charts/free5gc-amf free5gcns
```
Note: 
1. If <path_to_local_helm_chart> is not provided, then by default it would take the helm_charts present in Input-folder.

The generated Go-Code would be written to the "outputs/generated_code.go" file

The Generated Go-Code shall contain the following plugable functions:
1. Create_All():  When called, it will create all the k8s resources(services, deployment) on the kubernetes cluster.
2. Delete_All(): When called, it will delete all the k8s resources(services, deployment) on the kubernetes cluster.
3. Get_Resources(): Shall return the list of a particular resource.
    1. Get_Service(): Shall return the list of all services.
    2. Get_Deployment(): Shall return the list of all deployments. & so on

Further Docs:
1. Design Document: [link](https://docs.google.com/document/d/1b7WpK_BHe7nRuGP5MOy6Mxf3hpN_cro9/edit)
2. Detailed Algorithm: [link](https://1drv.ms/p/s!AkgeY1fT2A5UhQK4IWBxOJ6YUerh?e=BmBkRc)
