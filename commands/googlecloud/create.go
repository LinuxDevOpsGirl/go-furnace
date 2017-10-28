package googlecloud

import (
	"log"

	config "github.com/Skarlso/go-furnace/config/common"
	fc "github.com/Skarlso/go-furnace/config/google"
	"github.com/Skarlso/go-furnace/utils"
	"github.com/Yitsushi/go-commander"
	"github.com/fatih/color"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	dm "google.golang.org/api/deploymentmanager/v2"
	yaml "gopkg.in/yaml.v1"
)

// Create commands for google Deployment Manager
type Create struct {
}

// These need a better place
var keyName = color.New(color.FgWhite, color.Bold).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()

// Execute runs the create command
func (c *Create) Execute(opts *commander.CommandHelper) {
	log.Println("Creating Deployment under project name: .", keyName(fc.GOOGLEPROJECTNAME))
	deploymentName := config.STACKNAME
	log.Println("Deployment name is: ", keyName(deploymentName))
	ctx := context.Background()
	client, err := google.DefaultClient(ctx, dm.NdevCloudmanScope)
	if err != nil {
		log.Fatalf(err.Error())
	}
	d, _ := dm.New(client)
	deployments := constructDeploymen(deploymentName)
	ret := d.Deployments.Insert(fc.GOOGLEPROJECTNAME, deployments)
	_, err = ret.Do()
	if err != nil {
		log.Fatal("error while doing deployment: ", err)
	}
	utils.WaitForDeploymentToFinish(*d, deploymentName)
}

// Path contains all the jinja imports in the config.yml file.
type Path struct {
	Path string `yaml:"path"`
}

// Imports is the high level representation of imports in the config.yml file.
type Imports struct {
	Paths []Path `yaml:"imports"`
}

func constructDeploymen(deploymentName string) *dm.Deployment {
	gConfig := fc.LoadGoogleStackConfig()
	config := dm.ConfigFile{
		Content: string(gConfig),
	}
	targetConfiguration := dm.TargetConfiguration{
		Config: &config,
	}

	imps := Imports{}
	err := yaml.Unmarshal(gConfig, &imps)
	if err != nil {
		utils.HandleFatal("error while parsing yaml: ", err)
	}

	// Load templates and all .schema files that might accompany them.
	if len(imps.Paths) > 0 {
		imports := []*dm.ImportFile{}
		for _, temp := range imps.Paths {
			templateContent := fc.LoadImportFileContent(temp.Path)
			imports = append(imports, &dm.ImportFile{Content: string(templateContent)})
			if ok, schema := fc.LoadSchemaForPath(temp.Path); ok {
				imports = append(imports, &dm.ImportFile{Content: string(schema)})
			}
		}
		targetConfiguration.Imports = imports
	}

	deployments := dm.Deployment{
		Name:   deploymentName,
		Target: &targetConfiguration,
	}
	return &deployments
}

// NewCreate Creates a new create command
func NewCreate(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Create{},
		Help: &commander.CommandDescriptor{
			Name:             "create",
			ShortDescription: "Create a Google Deployment Manager",
			LongDescription:  `I'll write this later`,
			Arguments:        "",
			Examples:         []string{"create"},
		},
	}
}
