package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"strings"
	"flag"
	"github.com/google/cloudprober/probes/external/serverutils"
	"log"
	"github.com/golang/protobuf/proto"
)

var isserver = flag.Bool("server", false, "Whether to run in server mode")

func dockerProbe() (string, error) {
	var payload []string
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "",err
	}

	info, err := cli.Info(context.Background())
	if err != nil {
		return "",err
	}
	diskUsage, err := cli.DiskUsage(context.Background())
	if err != nil {
		return "",err
	}
	payload = append(payload, fmt.Sprintf("docker_layers_total_size %d", diskUsage.LayersSize))
	payload = append(payload, fmt.Sprintf("docker_images_count %d", len(diskUsage.Images)))
	payload = append(payload, fmt.Sprintf("docker_container_total_count %d", info.Containers))
	payload = append(payload, fmt.Sprintf("docker_containers_running_count %d", info.ContainersRunning))
	//payload = append(payload, fmt.Sprintf("success %t", true))

	return fmt.Sprintf(strings.Join(payload, "\n")),nil
}

func main() {
	flag.Parse()
	if *isserver {
		serverutils.Serve(func(request *serverutils.ProbeRequest, reply *serverutils.ProbeReply) {
			payload, err := dockerProbe()
			reply.Payload = proto.String(payload)
			if err != nil {
				reply.ErrorMessage = proto.String(err.Error())
			}
		})
	}
	payload, err := dockerProbe()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(payload)
}
