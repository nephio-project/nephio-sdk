# Helm to Operator Codegen Sdk
The "Helm to Operator Codegen Sdk" takes the helm chart as input and generates the golang-code which can be used by kubernetes operator to create/delete all the resources previously managed by the helm charts. The sdk can be employed to transition from helm-way-of-deploying-resources to the operator way.

Note: It is currently an experimental feature.

### Step 0: Prerequisite
1. GoLang Version: 1.21
2. Helm : v3.9.3
3. Go Packages:
```
# Clone the Repo
cd nephio-sdk/helm-to-operator-codegen-sdk/
go mod tidy
```

### Step 1: Running the sdk
```
go run main.go <path_to_local_helm_chart> <namespace> <logging-level>
```
Note:
1. The logging-level can be set to one of the following values: debug, info (default), error, warn
2. If <path_to_local_helm_chart> is not provided, then by default it would take the helm_charts present in Input-folder.

#### Example Run 
```
go run main.go /home/ubuntu/free5gccharts/towards5gs-helm/charts/free5gc/charts/free5gc-amf/ free5gcns info
```
<details>
<summary>The output is similar to:</summary>

```console
INFO[0000]  ----------------- Converting Helm to Yaml --------------------------
WARN[0000] Duplication Detected in Struct Mapping | For Preconditions
WARN[0000] Duplication Detected in Struct Mapping | For ConditionStatus
WARN[0000] Duplication Detected in Enum Mapping | For ConditionStatus
INFO[0000] CurFile --> | temp/templated/free5gc-amf/templates/amf-configmap.yaml
INFO[0000]  Current KRM Resource| Kind : ConfigMap| YamlFilePath : temp/templated/free5gc-amf/templates/amf-configmap.yaml
INFO[0000]       Converting Runtime to Json Completed
INFO[0000]       Converting Json to String Completed
INFO[0000] CurFile --> | temp/templated/free5gc-amf/templates/amf-deployment.yaml
INFO[0000]  Current KRM Resource| Kind : Deployment| YamlFilePath : temp/templated/free5gc-amf/templates/amf-deployment.yaml
INFO[0000]       Converting Runtime to Json Completed
INFO[0000]       Converting Json to String Completed
INFO[0000] CurFile --> | temp/templated/free5gc-amf/templates/amf-hpa.yaml
ERRO[0000] Unable to convert yaml to unstructured |Object 'Kind' is missing in 'null'
INFO[0000] CurFile --> | temp/templated/free5gc-amf/templates/amf-ingress.yaml
ERRO[0000] Unable to convert yaml to unstructured |Object 'Kind' is missing in 'null'
INFO[0000] CurFile --> | temp/templated/free5gc-amf/templates/amf-n2-nad.yaml
INFO[0000] Kind | NetworkAttachmentDefinition Would Be Treated as Third Party Kind
INFO[0000]       Converting Unstructured to String Completed
INFO[0000] CurFile --> | temp/templated/free5gc-amf/templates/amf-service.yaml
INFO[0000]  Current KRM Resource| Kind : Service| YamlFilePath : temp/templated/free5gc-amf/templates/amf-service.yaml
INFO[0000]       Converting Runtime to Json Completed
INFO[0000]       Converting Json to String Completed
INFO[0000] ----------------- Writing GO Code ---------------------------------
INFO[0000] ----------------- Program Run Successful| Summary ---------------------------------
INFO[0000] Deployment            |1
INFO[0000] NetworkAttachmentDefinition           |1
INFO[0000] Service               |1
INFO[0000] ConfigMap             |1
```
</details>


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
