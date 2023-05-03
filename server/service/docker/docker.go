package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/gorilla/websocket"
)

type DockerService struct{}

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

func GetImagesList() ([]types.ImageSummary, error) {
	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return nil, err
	}
	return images, nil
}

func GetContainerList() ([]types.Container, error) {

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}
	return containers, nil
}

type ContainerWithStats struct {
	Id          string
	CPUUsage    float64
	MemoryUsage uint64
	MemoryLimit uint64
}

func GetContainerStats(containerID string) (ContainerWithStats, error) {
	statsReader, err := cli.ContainerStats(context.Background(), containerID, false)
	if err != nil {
		return ContainerWithStats{}, err
	}

	var containerStats types.StatsJSON
	err = json.NewDecoder(statsReader.Body).Decode(&containerStats)

	cpuUsage := calculateCPUPercentage(&containerStats)
	memoryUsage := containerStats.MemoryStats.Usage
	memoryLimit := containerStats.MemoryStats.Limit

	var containerWithStats = ContainerWithStats{
		Id:          containerID,
		CPUUsage:    cpuUsage,
		MemoryUsage: memoryUsage,
		MemoryLimit: memoryLimit,
	}
	return containerWithStats, nil
}

func calculateCPUPercentage(stats *types.StatsJSON) float64 {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage) - float64(stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage) - float64(stats.PreCPUStats.SystemUsage)

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		return (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}

	return 0.0
}

func StartContainer(containerID string) error {
	err := cli.ContainerStart(context.Background(), containerID, types.ContainerStartOptions{})
	return err
}
func StopContainer(containerID string) error {
	err := cli.ContainerStop(context.Background(), containerID, container.StopOptions{})
	return err
}

func RemoveContainer(containerID string) error {
	removeOptions := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}
	err := cli.ContainerRemove(context.Background(), containerID, removeOptions)
	if err != nil {
		return err
	}

	return nil
}

func CreateAnacondaContainer(jupyterPort int, sshPort int) error {
	if !isPortAvailable(jupyterPort) {
		return fmt.Errorf("Jupyter port %d is not available", jupyterPort)
	}

	if !isPortAvailable(sshPort) {
		return fmt.Errorf("SSH port %d is not available", sshPort)
	}

	ctx := context.Background()

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			"8888/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%d", jupyterPort)}},
			"22/tcp":   []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%d", sshPort)}},
		},
		// Binds: []string{"/opt/notebook/workspace:/workspace"},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: "/opt/notebook/workspace",
				Target: "/workspace",
			},
		},
		Runtime: "nvidia",
	}

	config := &container.Config{
		Image: "yanfei/anaconda3:latest",
		Cmd: []string{
			"/bin/bash",
			"-c",
			"/usr/sbin/sshd && mkdir -p /workspace && jupyter notebook --NotebookApp.password='sha1:77b5117ca0a9:f62234b17bee56b22db9d5d2b307b7c42573569f' --notebook-dir=/workspace --ip='*' --port=8888 --no-browser --allow-root",
		},
		ExposedPorts: nat.PortSet{
			"8888/tcp": struct{}{},
			"22/tcp":   struct{}{},
		},
	}

	networkingConfig := &network.NetworkingConfig{}

	resp, err := cli.ContainerCreate(ctx, config, hostConfig, networkingConfig, nil, "")
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	fmt.Printf("Anaconda container started with ID: %s\n", resp.ID)
	return nil
}

func isPortAvailable(port int) bool {
	address := fmt.Sprintf(":%d", port)
	conn, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func MakeWebSocketConnection(conn *websocket.Conn, containerID string) {
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
