package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	// "time"
	"strings"
	"os"

	// "io/ioutil"
	// "log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

var client *ec2.Client

// EC2CreateInstanceAPI defines the interface for the RunInstances and CreateTags functions.
// We use this interface to test the functions using a mocked service.
type EC2CreateInstanceAPI interface {
	RunInstances(ctx context.Context,
		params *ec2.RunInstancesInput,
		optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)

	CreateTags(ctx context.Context,
		params *ec2.CreateTagsInput,
		optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
}

type Config struct {
	InstanceType string `json:"instance-type"`
	ImageId      string `json:"image-id"`
}

// MakeInstance creates an Amazon Elastic Compute Cloud (Amazon EC2) instance.
// Inputs:
//
//	c is the context of the method call, which includes the AWS Region.
//	api is the interface that defines the method call.
//	input defines the input arguments to the service call.
//
// Output:
//
//	If success, a RunInstancesOutput object containing the result of the service call and nil.
//	Otherwise, nil and an error from the call to RunInstances.
func MakeInstance(c context.Context, api EC2CreateInstanceAPI, input *ec2.RunInstancesInput) (*ec2.RunInstancesOutput, error) {
	return api.RunInstances(c, input)
}

// MakeTags creates tags for an Amazon Elastic Compute Cloud (Amazon EC2) instance.
// Inputs:
//
//	c is the context of the method call, which includes the AWS Region.
//	api is the interface that defines the method call.
//	input defines the input arguments to the service call.
//
// Output:
//
//	If success, a CreateTagsOutput object containing the result of the service call and nil.
//	Otherwise, nil and an error from the call to CreateTags.
func MakeTags(c context.Context, api EC2CreateInstanceAPI, input *ec2.CreateTagsInput) (*ec2.CreateTagsOutput, error) {
	return api.CreateTags(c, input)
}

type EC2TerminateInstanceAPI interface {
	TerminateInstances(ctx context.Context,
		params *ec2.TerminateInstancesInput,
		optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)
}

func DeleteInstance(c context.Context, api EC2TerminateInstanceAPI, input *ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error) {
	return api.TerminateInstances(c, input)
}

func DeleteInstancesCmd(name *string, value *string) {

	var instances = make([]string, 0)

	fmt.Println("Deleting instances with " + *name + "= " + *value)

	val := strings.Split(*value, ",")
	tag := "tag:" + *name

	input1 := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String(tag),
				Values: val,
			},
		},
	}

	descResult, err := client.DescribeInstances(context.TODO(), input1)
	if err != nil {
		fmt.Println("Got an error fetching the status of the instance")
		fmt.Println(err)
		return
	}

	for _, r := range descResult.Reservations {
		fmt.Println("Instance IDs:")
		for _, i := range r.Instances {
			//value := *i.InstanceId
			instances = append(instances, *i.InstanceId)
		}
		fmt.Println(instances)
	}

	input := &ec2.TerminateInstancesInput{
		InstanceIds: instances,
		DryRun:      new(bool),
	}

	delResult, err := DeleteInstance(context.TODO(), client, input)
	if err != nil {
		fmt.Println("Got an error terminating the instance:")
		fmt.Println(err)
		return
	}

	fmt.Println("Terminated instance with id: ", *delResult.TerminatingInstances[0].InstanceId)
}

func CreateInstancesCmd(name *string, value *string) {
	// Create separate values if required.
	minMaxCount := int32(1)

	// data, err := ioutil.ReadFile("/etc/config/config.json")
	// if err != nil {
	// 	log.Fatalf("Error reading config file: %v", err)
	// }

	// var config Config
	// if err := json.Unmarshal(data, &config); err != nil {
	// 	log.Fatalf("Error unmarshaling config: %v", err)
	// }

	// var insType types.InstanceType
	// insType = types.InstanceType(config.InstanceType)
	// var imgId *string
	// imgId = &config.ImageID

	var config Config

	file, err := os.Open("config/config.json")
	if err != nil {
		fmt.Println("Error opening config file:", err)
		os.Exit(1)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		fmt.Println("Error decoding config:", err)
		os.Exit(1)
	}

	fmt.Println("Configmap data:", config.ImageId, config.InstanceType)

	input := &ec2.RunInstancesInput{
		// ImageId:      aws.String("ami-0d0ca2066b861631c"),
		// InstanceType: types.InstanceTypeT2Micro,
		ImageId:      aws.String(config.ImageId),
		InstanceType: (types.InstanceType)(config.InstanceType),
		MinCount:     &minMaxCount,
		MaxCount:     &minMaxCount,
	}

	result, err := MakeInstance(context.TODO(), client, input)
	if err != nil {
		fmt.Println("Got an error creating an instance:")
		fmt.Println(err)
		return
	}

	tagInput := &ec2.CreateTagsInput{
		Resources: []string{*result.Instances[0].InstanceId},
		Tags: []types.Tag{
			{
				Key:   	name,
				Value: 	value,
			},
		},
	}

	_, err = MakeTags(context.TODO(), client, tagInput)
	if err != nil {
		fmt.Println("Got an error tagging the instance:")
		fmt.Println(err)
		return
	}

	fmt.Println("Created tagged instance with ID " + *result.Instances[0].InstanceId)
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}
	client = ec2.NewFromConfig(cfg)

}

func main() {
	fmt.Println("Provisioning/De-provisioning EC2 in progress")
	command := flag.String("c", "", "command  create or delete")
	name := flag.String("n", "", "The name of the tag to attach to the instance")
	value := flag.String("v", "", "The value of the tag to attach to the instance")

	flag.Parse()

	// time.Sleep(5 * time.Minute)

	if *command == "create" {
		if *name == "" || *value == "" {
			fmt.Println("You must supply a name and value for the tag (-n TagName -v TagValue)")
			return
		}
		CreateInstancesCmd(name, value)
	} else if *command == "delete" {
		if *name == "" || *value == "" {
			fmt.Println("You must supply a name and value for the tag (-n TagName -v TagValue)")
			return
		}
		DeleteInstancesCmd(name, value)
	} else {
		fmt.Println("You must supply an command create or delete (-c create")
		return
	}
}
// End