package commands

import (
	"fmt"
	"log"
	"os"

	config "github.com/Skarlso/go-furnace/config"
	awsconfig "github.com/Skarlso/go-furnace/furnace-aws/config"
	"github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/fatih/color"
)

// Status command.
type Status struct {
}

// ResourceStatus defines a resource.
type ResourceStatus struct {
	// Status is the status of the resource
	Status cloudformation.ResourceStatus
	// PhysicalId of the resource
	PhysicalID string
	// LogicalId of the resource
	LogicalID string
	// Type of the resource
	Type string
}

// Execute defines what this command does.
func (c *Status) Execute(opts *commander.CommandHelper) {
	configName := opts.Arg(0)
	if len(configName) > 0 {
		dir, _ := os.Getwd()
		if err := awsconfig.LoadConfigFileIfExists(dir, configName); err != nil {
			config.HandleFatal(configName, err)
		}
	}
	stackname := awsconfig.Config.Main.Stackname
	cfg, err := external.LoadDefaultAWSConfig()
	config.CheckError(err)
	cfClient := cloudformation.New(cfg)
	client := CFClient{cfClient}
	stack := stackStatus(stackname, &client)
	info := color.New(color.FgWhite, color.Bold).SprintFunc()
	log.Println("Stack state is: ", info(stack.Stacks[0].GoString()))
	stackResources := stackResources(stackname, &client)
	printStackResources(stackResources)
}

func stackStatus(stackname string, cfClient *CFClient) *cloudformation.DescribeStacksOutput {
	req := cfClient.Client.DescribeStacksRequest(&cloudformation.DescribeStacksInput{StackName: aws.String(stackname)})
	descResp, err := req.Send()
	config.CheckError(err)
	fmt.Println()
	return descResp
}

func stackResources(stackname string, cfClient *CFClient) []ResourceStatus {
	resources := make([]ResourceStatus, 0)
	req := cfClient.Client.DescribeStackResourcesRequest(&cloudformation.DescribeStackResourcesInput{StackName: aws.String(stackname)})
	descResp, err := req.Send()
	config.CheckError(err)
	for _, r := range descResp.StackResources {
		res := ResourceStatus{Status: r.ResourceStatus, PhysicalID: *r.PhysicalResourceId, LogicalID: *r.LogicalResourceId, Type: *r.ResourceType}
		resources = append(resources, res)
	}
	fmt.Println()
	return resources
}

func printStackResources(resources []ResourceStatus) {
	info := color.New(color.FgWhite, color.Bold).SprintFunc()
	fmt.Println(info("___________________"))
	for _, r := range resources {
		fmt.Print(info(r))
	}
	fmt.Println()
}

func (r ResourceStatus) String() string {
	var red = color.New(color.FgRed).SprintFunc()
	var yellow = color.New(color.FgYellow).SprintFunc()
	ret := ""
	ret += fmt.Sprintf("|Name:          %s|\n|Id:            %s|\n|Status:        %s|\n|Type:          %s|\n",
		red(r.LogicalID),
		yellow(r.PhysicalID),
		yellow(r.Status),
		yellow(r.Type))
	ret += "-------------------\n"
	return ret
}

// NewStatus Creates a new Status command.
func NewStatus(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Status{},
		Help: &commander.CommandDescriptor{
			Name:             "status",
			ShortDescription: "Status of a stack.",
			LongDescription:  `Get detailed status of the stack.`,
			Arguments:        "custom-config",
			Examples:         []string{"", "custom-config"},
		},
	}
}
