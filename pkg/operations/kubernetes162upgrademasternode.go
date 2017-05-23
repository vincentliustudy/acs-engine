package operations

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Azure/acs-engine/pkg/api"
	"github.com/Azure/acs-engine/pkg/armhelpers"
	log "github.com/Sirupsen/logrus"
)

// Compiler to verify QueueMessageProcessor implements OperationsProcessor
var _ UpgradeNode = &UpgradeMasterNode{}

// UpgradeMasterNode upgrades a Kubernetes 1.5.3 master node to 1.6.2
type UpgradeMasterNode struct {
	TemplateMap             map[string]interface{}
	ParametersMap           map[string]interface{}
	UpgradeContainerService *api.ContainerService
	ResourceGroup           string
	Client                  armhelpers.ACSEngineClient
}

// DeleteNode takes state/resources of the master/agent node from ListNodeResources
// backs up/preserves state as needed by a specific version of Kubernetes and then deletes
// the node
func (kmn *UpgradeMasterNode) DeleteNode(vmName *string) error {
	if err := CleanDeleteVirtualMachine(kmn.Client, kmn.ResourceGroup, *vmName); err != nil {
		log.Fatalln(err)
		return err
	}

	return nil
}

// CreateNode creates a new master/agent node with the targeted version of Kubernetes
func (kmn *UpgradeMasterNode) CreateNode(poolName string, masterOffset int) error {
	templateVariables := kmn.TemplateMap["variables"].(map[string]interface{})

	// Call CreateVMWithRetries
	templateVariables["masterOffset"] = masterOffset
	masterOffsetVar, _ := templateVariables["masterOffset"]
	log.Infoln(fmt.Sprintf("Master offset: %v", masterOffsetVar))

	WriteTemplate(kmn.UpgradeContainerService, kmn.TemplateMap, kmn.ParametersMap)

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	deploymentSuffix := random.Int31()

	_, err := kmn.Client.DeployTemplate(
		kmn.ResourceGroup,
		fmt.Sprintf("%s-%d", kmn.ResourceGroup, deploymentSuffix),
		kmn.TemplateMap,
		kmn.ParametersMap,
		nil)

	if err != nil {
		log.Fatalln(err)
	}

	return nil
}

// Validate will verify the that master/agent node has been upgraded as expected.
func (kmn *UpgradeMasterNode) Validate() error {
	return nil
}
