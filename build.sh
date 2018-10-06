protoc -I pkg/govox/ pkg/govox/govox.proto --go_out=plugins=grpc:pkg/govox
python -m grpc_tools.protoc -I pkg/govox/ pkg/govox/govox.proto --python_out=services/customgen --grpc_python_out=services/customgen
