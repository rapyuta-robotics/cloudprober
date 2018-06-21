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
	epb "github.com/google/cloudprober/probes/external/proto"
	"strconv"
)

var isserver = flag.Bool("server", false, "Whether to run in server mode")

const (
	BYTE     = 1.0 << (10 * iota)
	KILOBYTE
	MEGABYTE
	GIGABYTE
	TERABYTE
)

var interestingKeys = map[string]string{
	"Pool Blocksize":               "pool_blocksize",
	"Base DeviceSize":              "base_device_size",
	"Data Space Used":              "data_space_used",
	"Data Space Total":             "data_space_total",
	"Data Space Available":         "data_space_available",
	"Metadata Space Used":          "metadata_space_used",
	"Metadata Space Total":         "metadata_space_total",
	"Metadata Space Available":     "metadata_space_available",
	"Thin Pool Minimum Free Space": "thin_pool_minimum_free_space",
}

func dockerProbe() (string, error) {
	var payload []string
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", err
	}

	info, err := cli.Info(context.Background())
	if err != nil {
		return "", err
	}
	for _, keyValTup := range info.DriverStatus {
		promKey, ok := interestingKeys[keyValTup[0]]
		if ok {
			value, err := convertToBytes(keyValTup[1])
			if err == nil {
				payload = append(payload, fmt.Sprintf("%s %d", promKey, value))
			}
		}
	}
	payload = append(payload, fmt.Sprintf("container_total %d", info.Containers))
	payload = append(payload, fmt.Sprintf("containers_running %d", info.ContainersRunning))
	payload = append(payload, fmt.Sprintf("images %d", info.Images))
	//payload = append(payload, fmt.Sprintf("success %t", true))

	return fmt.Sprintf(strings.Join(payload, "\n")), nil
}

func convertToBytes(input string) (int64, error) {
	frags := strings.Split(input, " ")
	valStr, unit := frags[0], strings.ToLower(frags[1])
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return 0, err
	}
	multiplier := 1
	switch {
	case unit == "kb":
		multiplier = KILOBYTE
	case unit == "mb":
		multiplier = MEGABYTE
	case unit == "gb":
		multiplier = GIGABYTE
	case unit == "tb":
		multiplier = TERABYTE
	}
	return int64(val * float64(multiplier)), nil

}

func main() {
	flag.Parse()
	if *isserver {
		serverutils.Serve(func(request *epb.ProbeRequest, reply *epb.ProbeReply) {
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
