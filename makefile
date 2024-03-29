buildVersion = $$(git describe --tags)
buildDate = $$(date)
buildCommit = $$(git rev-parse HEAD)

agent:
	cd cmd/agent && \
	go build -v -ldflags="-X 'main.buildVersion=$(buildVersion)' -X 'main.buildDate=$(buildDate)' -X main.buildCommit=$(buildCommit)" 

server:
	cd cmd/server && \
	go build -v -ldflags="-X 'main.buildVersion=$(buildVersion)' -X 'main.buildDate=$(buildDate)' -X main.buildCommit=$(buildCommit)" 

pb :
	protoc --proto_path=proto proto/*.proto --go_out=internal --go-grpc_out=internal