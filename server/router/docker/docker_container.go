package docker

import (
	v1 "github.com/flipped-aurora/gin-vue-admin/server/api/v1"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type DockerRouter struct{}

func (e *DockerRouter) InitWsRouter(Router *gin.RouterGroup) {
	dockerRouterWithoutRecord := Router.Group("docker").Use()
	dockerApi := v1.ApiGroupApp.DockerAPiGroup.DockerApi
	dockerRouterWithoutRecord.GET("/containers/:id/console", dockerApi.WsHandler)
}

func (e *DockerRouter) InitDockerRouter(Router *gin.RouterGroup) {
	dockerRouter := Router.Group("docker").Use(middleware.OperationRecord())
	dockerApi := v1.ApiGroupApp.DockerAPiGroup.DockerApi
	{
		dockerRouter.POST("containers", dockerApi.GetContainerListHandler)
		dockerRouter.POST("/containers/:id/start", dockerApi.StartContainerHandler)
		dockerRouter.POST("/containers/:id/stop", dockerApi.StopContainerHandler)
		dockerRouter.POST("/containers/:id/remove", dockerApi.RemoveContainerHandler)
		dockerRouter.POST("/containers/:id/stats", dockerApi.GetContainerStatsHandler)
		dockerRouter.POST("/create_anaconda_container", dockerApi.CreateAnacondaContainerHandler)
	}

}
