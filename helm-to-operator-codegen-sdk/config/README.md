# Helm to Operator Codegen Sdk 
## Config
The config-files consists of two jsons:
1. Struct Module Mapping : It Maps the structs (types) of the package with its package-name. ("Package-Name": [List of all the structs it contain])
2. Enum Module Mapping: It Maps the enums of the package with its package-name. ("Package-Name": [List of all the enums it contain])

## Significance
The above config files solves the following problems:

### Problem-1: “Which v1 Package?”
Mostly, It is seen that inspecting the type of struct(using reflect) would tell us that the struct belong to package “v1”, but there are multiple v1 packages (appsv1, metav1, rbacv1 etc), So, the actual package remains unknown. 

Solution: In order to solve above problems, we build a “structModuleMapping” which is a map that takes “struct-name” as key and gives “package/module name” as value.
```
    v1.Deployment  -->  appsv1.Deployment
    v1.Service     --> corev1.Service
```

### Problem-2: “Data-Type is Struct or Enum?”
Structs needs to be initialised using curly-brackets {}, whereas enums needs Paranthesis (), Since, reflect doesn’t tell us which data-type is struct or enum, We:

Solution: We solve above problems by building a “enumModuleMapping” which is a set that stores all data-types that are enums. i.e. If a data-type belongs to the set, then It is a Enum.
