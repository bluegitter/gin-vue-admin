package docker

import (
	"context"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/flipped-aurora/gin-vue-admin/server/service/docker"
	"github.com/gorilla/websocket"

	"github.com/gin-gonic/gin"
)

type DockerApi struct{}

func (e *DockerApi) GetImagesListHandler(c *gin.Context) {
	images, err := docker.GetImagesList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -1,
			"data": gin.H{},
			"msg":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": images,
		"msg":  "Container started",
	})
}

// ContainerListHandler handles the /containers API request.
func (e *DockerApi) GetContainerListHandler(c *gin.Context) {
	containerList, err := docker.GetContainerList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -1,
			"data": gin.H{},
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": containerList,
		"msg":  "Container started",
	})
}

func (e *DockerApi) GetContainerStatsHandler(c *gin.Context) {
	containerID := c.Param("id")
	containerStatsList, err := docker.GetContainerStats(containerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -1,
			"data": gin.H{},
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": containerStatsList,
		"msg":  "Container started",
	})
}

func (e *DockerApi) StartContainerHandler(c *gin.Context) {
	containerID := c.Param("id")
	err := docker.StartContainer(containerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -1,
			"data": gin.H{},
			"msg":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{},
		"msg":  "Container started",
	})
}

func (e *DockerApi) StopContainerHandler(c *gin.Context) {
	containerID := c.Param("id")
	err := docker.StopContainer(containerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -1,
			"data": gin.H{},
			"msg":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{},
		"msg":  "Container stoped",
	})
}

func (e *DockerApi) RemoveContainerHandler(c *gin.Context) {
	containerID := c.Param("id")
	err := docker.RemoveContainer(containerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -1,
			"data": gin.H{},
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{},
		"msg":  "Container removed successfully",
	})
}

func (e *DockerApi) CreateAnacondaContainerHandler(c *gin.Context) {
	jupyterPortStr := c.PostForm("jupyter_port")
	sshPortStr := c.PostForm("ssh_port")

	jupyterPort, err := strconv.Atoi(jupyterPortStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -1,
			"data": gin.H{},
			"msg":  "Invalid Jupyter port",
		})
		return
	}

	sshPort, err := strconv.Atoi(sshPortStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -1,
			"data": gin.H{},
			"msg":  "Invalid SSH port",
		})
		return
	}
	err = docker.CreateAnacondaContainer(jupyterPort, sshPort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{},
		"msg":  "Anaconda container created.",
	})
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var cli *client.Client

func init() {
	var err error
	// 创建Docker客户端
	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithVersion("1.41"))
	if err != nil {
		log.Fatalf("Error creating Docker client: %v", err)
		return
	}
}

func (e *DockerApi) WsHandler(c *gin.Context) {
	containerID := c.Param("id")

	if containerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "containerID is required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer conn.Close()

	execConfig := types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{"/bin/bash"},
	}

	exec, err := cli.ContainerExecCreate(context.Background(), containerID, execConfig)
	if err != nil {
		log.Fatal(err)
	}

	attach := types.ExecStartCheck{
		Tty: true,
	}

	resp, err := cli.ContainerExecAttach(context.Background(), exec.ID, attach)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Close()

	go func() {
		for {
			_, r, err := conn.NextReader()
			if err != nil {
				break
			}
			_, err = io.Copy(resp.Conn, r)
			if err != nil {
				break
			}
		}
	}()

	buf := make([]byte, 1024)
	for {
		n, err := resp.Reader.Read(buf)
		if err != nil {
			break
		}
		err = conn.WriteMessage(websocket.TextMessage, buf[:n])
		if err != nil {
			break
		}
	}
}
