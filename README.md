# helm-to-operator-codegen-sdk 
### Prerequisties:
1. GoLang Version: 1.19
```
wget -c https://golang.org/dl/go1.19.8.linux-amd64.tar.gz
sudo tar -C /usr/local -xvzf go1.19.8.linux-amd64.tar.gz
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
git clone https://github.sec.samsung.net/s-jaisawal/helm-operatort-sdk.git
cd helm-operatort-sdk/
go mod tidy
```

### Running the Script
```
go run main.go <path_to_local_helm_chart> <namespace>
e.g. go run main.go /home/ubuntu/free5gccharts/towards5gs-helm/charts/free5gc/charts/free5gc-amf free5gcns
```
Note: 
1. If <path_to_local_helm_chart> is not provided, then by default it would take the helm_charts present in Input-folder.

The generated Go-Code would be written in the "outputs/generated_code.go" file

The Generated Go-Code would contain the following plugable function:
1. Create_All():  It when called, create all the k8s resources(services, deployment) in the kubernetes cluster.
2. Get_Resources(): would return the list of a particular resource.
    1. Get_Service() would return the list of all Services-Objects.
    2. Get_Deployment() would return the list of all Deployment-Objects. & so on
