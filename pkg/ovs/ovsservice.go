package ovs

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	plsv1 "github.com/Networks-it-uc3m/l2sm-switch/api/v1"
)

type OvsService struct {
	exec Client
}

const NO_DEFAULT_ID = -1

func NewOvsService() OvsService {
	return OvsService{exec: NewClient(OvsVsctlClient)}
}

func NewSudoOvsService() OvsService {
	return OvsService{exec: NewSudoClient(OvsVsctlClient)}
}

func (ovsService *OvsService) AddBridge(bridgeName string) error {
	output, err := ovsService.exec.CombinedOutput("add-br", bridgeName)
	if err != nil {
		return fmt.Errorf("add-br error: %v\nOutput: %s", err, output)

	}
	return nil
}

func (ovsService *OvsService) DeleteBridge(bridgeName string) error {
	output, err := ovsService.exec.CombinedOutput("del-br", bridgeName)
	if err != nil {
		return fmt.Errorf("del-br error: %v\nOutput: %s", err, output)

	}
	return nil
}

func (ovsService *OvsService) SetDatapathID(bridgeName, datapathId string) error {
	output, err := ovsService.exec.CombinedOutput("set", "bridge", bridgeName, fmt.Sprintf("other-config:datapath-id=%s", datapathId))
	if err != nil {
		return fmt.Errorf("set bridge error: %v\nOutput: %s", err, output)
	}
	return nil
}

func (ovsService *OvsService) SetProtocol(bridgeName, protocol string) error {
	protocolString := fmt.Sprintf("protocols=%s", protocol)

	output, err := ovsService.exec.CombinedOutput("set", "bridge", bridgeName, protocolString)
	if err != nil {
		return fmt.Errorf("set bridge error: %v\nOutput: %s", err, output)
	}
	return nil
}
func (ovsService *OvsService) SetController(bridgeName string, controller ...string) error {

	output, err := ovsService.exec.CombinedOutput("set-controller", bridgeName, strings.Join(controller, " "))
	if err != nil {
		return fmt.Errorf("set-controller error: %v\nOutput: %s", err, output)
	}

	return nil
}

func (ovsService *OvsService) CreateVxlan(bridgeName string, vxlan plsv1.Vxlan) error {
	commandArgs := []string{
		"add-port",
		bridgeName,
		vxlan.VxlanId,
		"--",
		"set", "interface",
		vxlan.VxlanId,
		"type=vxlan",
		"options:key=flow",
		fmt.Sprintf("options:remote_ip=%s", vxlan.RemoteIp),
		fmt.Sprintf("options:local_ip=%s", vxlan.LocalIp),
		fmt.Sprintf("options:dst_port=%s", vxlan.UdpPort),
	}
	output, err := ovsService.exec.CombinedOutput(commandArgs...)

	if err != nil {
		return fmt.Errorf("add-port error: %v\nOutput: %s", err, output)
	}
	return nil

}

func (ovsService *OvsService) DeleteVxlan(bridgeName, vxlanId string) error {
	commandArgs := []string{
		"del-port",
		bridgeName,
		vxlanId,
	}

	output, err := ovsService.exec.CombinedOutput(commandArgs...)
	if err != nil {
		return fmt.Errorf("del-port error: %v\nOutput: %s", err, output)
	}
	return nil
}

func (ovsService *OvsService) ModifyVxlan(vxlan plsv1.Vxlan) error {
	commandArgs := []string{
		"set", "interface",
		vxlan.VxlanId,
		"type=vxlan",
		"options:key=flow",
		fmt.Sprintf("options:remote_ip=%s", vxlan.RemoteIp),
		fmt.Sprintf("options:local_ip=%s", vxlan.LocalIp),
		fmt.Sprintf("options:dst_port=%s", vxlan.UdpPort),
	}
	output, err := ovsService.exec.CombinedOutput(commandArgs...)

	if err != nil {
		return fmt.Errorf("set interface error: %v\nOutput: %s", err, output)
	}
	return nil

}

func (ovsService *OvsService) AddPort(bridgeName, portName string, netIndex int, internal bool) error {
	args := []string{"add-port", bridgeName, portName}

	if netIndex != NO_DEFAULT_ID {
		args = append(args,
			"--",
			"set", "Interface", portName, fmt.Sprintf("ofport_request=%d", netIndex),
		)
	}
	if internal {
		args = append(args,
			"--",
			"set", "interface", portName, "type=internal")
	}

	output, err := ovsService.exec.CombinedOutput(args...)
	if err != nil {
		return fmt.Errorf("add-port error: %v\nOutput: %s", err, output)
	}

	return nil
}

// TODO: correct formats. Be careful because i dont remember what the outut of get interface was, so i need to check
// and pass it to integer or string depending on the situation
func (ovsService *OvsService) GetPortNumber(portName string) (int64, error) {
	output, err := ovsService.exec.CombinedOutput("get", "Interface", portName, "ofport")
	if err != nil {
		return 0, fmt.Errorf("get Interface error: %v\nOutput: %s", err, output)
	}

	ofportStr := strings.TrimSpace(string(output))
	ofport, err := strconv.ParseInt(ofportStr, 10, 64)
	return ofport, err
}

func (ovsService *OvsService) GetPorts(bridgeName string) (map[string]plsv1.Port, error) {
	portMap := make(map[string]plsv1.Port)
	output, err := ovsService.exec.CombinedOutput("list-ports", bridgeName)
	if err != nil {
		return portMap, fmt.Errorf("list-ports error: %v\nOutput: %s", err, output)
	}

	portNames := strings.Split(string(output), "\n") //TODO: Check
	for _, portName := range portNames {
		if portName == "" {
			continue
		}
		// TODO:, retrieve more details for each port; here we just set the name
		portMap[portName] = plsv1.Port{Name: portName}
		// Retrieve status
		// cmd = exec.Command("ovs-vsctl", "get", "Interface", portName, "status")

	}

	return portMap, nil
}

func (ovsService *OvsService) GetController(bridgeName string) ([]string, error) {
	controllers := []string{}
	output, err := ovsService.exec.CombinedOutput("get-controller", bridgeName)
	if err != nil {
		return controllers, fmt.Errorf("get-controller error: %v\nOutput: %s", err, output)
	}

	// Split the output by lines for each controller name
	controllersOutput := strings.Split(string(output), "\n") // TODO: CHECK
	for _, controller := range controllersOutput {
		if controller == "" {
			continue
		}
		controllers = append(controllers, controller)
	}

	return controllers, nil
}

// Define a helper struct to match the ovs-vsctl JSON output
type OVSVxlanOutput struct {
	Data     [][]any  `json:"data"`
	Headings []string `json:"headings"`
}

func (ovsService *OvsService) GetVxlans(bridgeName string) (map[string]plsv1.Vxlan, error) {

	vxlans := []plsv1.Vxlan{}

	output, err := ovsService.exec.CombinedOutput("--column=name,options", "--format=json", "--data=json", "find", "Interface", "type=vxlan")

	if err != nil {
		return map[string]plsv1.Vxlan{}, fmt.Errorf("find Interface type=vxlan error: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if outputStr == "" {
		// No VXLANs found, which is not necessarily an error.
		return map[string]plsv1.Vxlan{}, nil
	}

	var ovsOutput OVSVxlanOutput
	err = json.Unmarshal([]byte(outputStr), &ovsOutput)
	if err != nil {
		return map[string]plsv1.Vxlan{}, fmt.Errorf("failed to unmarshal ovs-vsctl JSON output: %v\nOutput: %s", err, outputStr)
	}

	// Ensure data is present and has the expected structure
	if len(ovsOutput.Data) == 0 {
		return map[string]plsv1.Vxlan{}, nil // No VXLAN interfaces found
	}

	for _, item := range ovsOutput.Data {
		if len(item) < 2 {
			// log.Printf("Skipping malformed item: %v", item) // Or handle error
			continue
		}

		vxlanName, ok := item[0].(string)
		if !ok {
			// log.Printf("Skipping item with non-string name: %v", item[0])
			continue
		}

		optionsList, ok := item[1].([]any)
		if !ok || len(optionsList) < 2 || optionsList[0] != "map" {
			// log.Printf("Skipping item with malformed options: %v", item[1])
			continue
		}

		optionsMapData, ok := optionsList[1].([]any)
		if !ok {
			// log.Printf("Skipping item with malformed options map data: %v", optionsList[1])
			continue
		}

		vxlan := plsv1.Vxlan{VxlanId: vxlanName}
		options := make(map[string]string)

		for _, optionPair := range optionsMapData {
			pair, ok := optionPair.([]any)
			if !ok || len(pair) < 2 {
				// log.Printf("Skipping malformed option pair: %v", optionPair)
				continue
			}
			key, keyOk := pair[0].(string)
			value, valueOk := pair[1].(string)
			if keyOk && valueOk {
				options[key] = value
			}
		}

		vxlan.LocalIp = options["local_ip"]
		vxlan.RemoteIp = options["remote_ip"]
		vxlan.UdpPort = options["dst_port"]
		// The "key" option is usually "flow", you can store it if needed:
		// vxlan.Key = options["key"]

		vxlans = append(vxlans, vxlan)
	}
	vxlansMap := map[string]plsv1.Vxlan{}
	for _, v := range vxlans {
		vxlansMap[v.VxlanId] = v
	}
	return vxlansMap, nil
}
