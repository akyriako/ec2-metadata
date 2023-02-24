package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

const metadataBaseEndpoint = "http://169.254.169.254/latest/meta-data/"

type Ec2Metadata struct {
	Hostname           string `json:"hostname,omitempty"`
	PrivateIpV4Address string `json:"privateIpV4Address,omitempty"`
	PublicIpV4Address  string `json:"publicIpV4Address,omitempty"`
	InstanceType       string `json:"instanceType,omitempty"`
	AvailabilityZone   string `json:"availabilityZone,omitempty"`
}

func main() {
	envPath := flag.String("path", "", "write metadata as an environment file to this path")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	ec2Metadata, err := getMetadata()
	if err != nil {
		log.Fatalln(err)
	}

	if envPath != nil && strings.Trim(*envPath, " ") != "" {
		err := writeEnvFile(ec2Metadata, envPath)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		ec2MetadataAsJson, err := json.Marshal(ec2Metadata)
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println(string(ec2MetadataAsJson))
	}
}

func getMetadata() (*Ec2Metadata, error) {
	hostname, err := getHostName()
	if err != nil {
		return nil, err
	}

	privateIp, err := getPrivateIpV4Address()
	if err != nil {
		return nil, err
	}

	publicIp, err := getPublicIpV4Address()
	if err != nil {
		return nil, err
	}

	instanceType, err := getInstanceType()
	if err != nil {
		return nil, err
	}

	availabilityZone, err := getAvailabilityZone()
	if err != nil {
		return nil, err
	}

	ecsMetadata := &Ec2Metadata{
		Hostname:           hostname,
		PrivateIpV4Address: privateIp,
		PublicIpV4Address:  publicIp,
		InstanceType:       instanceType,
		AvailabilityZone:   availabilityZone,
	}

	return ecsMetadata, nil
}

func getResponse(metadataUrlPath string) (string, error) {
	endpoint, _ := url.Parse(metadataBaseEndpoint)
	endpoint.Path = path.Join(endpoint.Path, metadataUrlPath)

	response, err := http.Get(endpoint.String())
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(responseBody), nil
}

func getHostName() (string, error) {
	value, err := getResponse("hostname")
	if err != nil {
		return "", err
	}

	return value, nil
}

func getPrivateIpV4Address() (string, error) {
	value, err := getResponse("local-ipv4")
	if err != nil {
		return "", err
	}

	return value, nil
}

func getPublicIpV4Address() (string, error) {
	value, err := getResponse("public-ipv4")
	if err != nil {
		return "", err
	}

	return value, nil
}

func getInstanceType() (string, error) {
	value, err := getResponse("instance-type")
	if err != nil {
		return "", err
	}

	return value, nil
}

func getAvailabilityZone() (string, error) {
	value, err := getResponse("placement/availability-zone")
	if err != nil {
		return "", err
	}

	return value, nil
}

func writeEnvFile(ec2Metadata *Ec2Metadata, path *string) error {
	envMap := make(map[string]string)

	envMap["META_EC2_HOSTNAME"] = ec2Metadata.Hostname
	envMap["META_EC2_INSTANCE_TYPE"] = ec2Metadata.InstanceType
	envMap["META_EC2_AVAILABILITY_ZONE"] = ec2Metadata.AvailabilityZone
	envMap["META_EC2_PRIVATE_IP"] = ec2Metadata.PrivateIpV4Address
	envMap["META_EC2_PUBLIC_IP"] = ec2Metadata.PublicIpV4Address

	//filename := filepath.Join(*path, "ec2-metadata")
	err := godotenv.Write(envMap, *path)

	return err
}
