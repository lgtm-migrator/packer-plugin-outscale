package common

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	oscgo "github.com/outscale/osc-sdk-go/v2"

	"context"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// Create statebag for running test
func getState() (multistep.StateBag, error) {
	state := new(multistep.BasicStateBag)
	accessConfig := &AccessConfig{}
	accessConfig.RawRegion = "eu-west-2"
	var oscConn *oscgo.APIClient
	var err error
	if oscConn, err = accessConfig.NewOSCClient(); err != nil {
		err := fmt.Errorf("error in creating osc Client: %s", err.Error())
		return nil, err
	}
	state.Put("osc", oscConn)
	state.Put("ui", &packersdk.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})
	state.Put("accessConfig", accessConfig)
	return state, err
}

func TestMostRecentOmiFilter(t *testing.T) {
	stepSourceOMIInfo := StepSourceOMIInfo{
		SourceOmi: "ami-7cab7c18",
		OmiFilters: OmiFilterOptions{
			MostRecent: true,
		},
	}
	state, err := getState()
	if state == nil {
		t.Fatalf("error retrieving state %s", err.Error())
	}

	action := stepSourceOMIInfo.Run(context.Background(), state)
	if err := state.Get("error"); err != nil {
		t.Fatalf("should not error, but: %v", err)
	}

	if action != multistep.ActionContinue {
		t.Fatalf("shoul continue, but: %v", action)
	}

}
