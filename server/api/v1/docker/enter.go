package docker

import "github.com/flipped-aurora/gin-vue-admin/server/service"

type ApiGroup struct {
	DockerApi
}

var (
	dockerService = service.ServiceGroupApp.DockerServiceGroup.DockerService
)
