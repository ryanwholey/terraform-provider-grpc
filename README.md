# terraform-provider-grpc

A Terraform provider for making gRPC requests

*Warning* This project is in an experimental phase and subject to change. 

## Usage

```hcl
data "grpc_request" "get_version_info" {
  address = "grpc-server.com:443"
  method  = "org.service.Service.GetVersionInfo"
}

data "grpc_request" "list_resources" {
  address = "grpc-server.com:443"
  method  = "org.service.Service.ListResources"

  request_headers = {
    "client-api-protocol" = "1,1"
    "authorization"       = var.auth_token
  }
}

output "requests" {
  value = {
    GetVersionInfo = jsondecode(data.grpc_request.get_version_info.body),
    ListProjects   = jsondecode(data.grpc_request.list_projects.body),
  }
}
```
