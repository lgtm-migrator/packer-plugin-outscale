package chroot

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	oscgo "github.com/outscale/osc-sdk-go/v2"
	osccommon "github.com/outscale/packer-plugin-outscale/builder/osc/common"
)

// StepCreateVolume creates a new volume from the snapshot of the root
// device of the OMI.
//
// Produces:
//
//	volume_id string - The ID of the created volume
type StepCreateVolume struct {
	volumeId       string
	RootVolumeSize int64
	RootVolumeType string
	RootVolumeTags osccommon.TagMap
	RawRegion      string
	Ctx            interpolate.Context
}

func (s *StepCreateVolume) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	oscconn := state.Get("osc").(*oscgo.APIClient)
	vm := state.Get("vm").(oscgo.Vm)
	ui := state.Get("ui").(packersdk.Ui)

	var err error

	volTags, err := s.RootVolumeTags.OSCTags(s.Ctx, s.RawRegion, state)

	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	var createVolume *oscgo.CreateVolumeRequest
	if config.FromScratch {
		rootVolumeType := osccommon.VolumeTypeGp2
		if s.RootVolumeType == "io1" {
			err := errors.New("Cannot use io1 volume when building from scratch")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		} else if s.RootVolumeType != "" {
			rootVolumeType = s.RootVolumeType
		}
		size := int32(s.RootVolumeSize)
		createVolume = &oscgo.CreateVolumeRequest{
			SubregionName: vm.Placement.GetSubregionName(),
			Size:          &size,
			VolumeType:    &rootVolumeType,
		}

	} else {
		// Determine the root device snapshot
		image := state.Get("source_image").(oscgo.Image)
		log.Printf("Searching for root device of the image (%s)", image.GetRootDeviceName())
		var rootDevice *oscgo.BlockDeviceMappingImage
		for _, device := range *image.BlockDeviceMappings {
			if device.DeviceName == image.RootDeviceName {
				rootDevice = &device
				break
			}
		}

		ui.Say("Creating the root volume...")
		createVolume, err = s.buildCreateVolumeInput(vm.Placement.GetSubregionName(), rootDevice)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	log.Printf("Create args: %+v", createVolume)

	createVolumeResp, _, err := oscconn.VolumeApi.CreateVolume(context.Background()).CreateVolumeRequest(*createVolume).Execute()
	if err != nil {
		err := fmt.Errorf("Error creating root volume: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the volume ID so we remember to delete it later
	s.volumeId = *createVolumeResp.GetVolume().VolumeId
	log.Printf("Volume ID: %s", s.volumeId)

	//Create tags for volume
	if len(volTags) > 0 {
		if err := osccommon.CreateOSCTags(oscconn, s.volumeId, ui, volTags); err != nil {
			err := fmt.Errorf("Error creating tags for volume: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// Wait for the volume to become ready
	err = osccommon.WaitUntilOscVolumeAvailable(oscconn, s.volumeId)
	if err != nil {
		err := fmt.Errorf("Error waiting for volume: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("volume_id", s.volumeId)
	return multistep.ActionContinue
}

func (s *StepCreateVolume) Cleanup(state multistep.StateBag) {
	if s.volumeId == "" {
		return
	}

	oscconn := state.Get("osc").(*oscgo.APIClient)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Deleting the created BSU volume...")
	_, _, err := oscconn.VolumeApi.DeleteVolume(context.Background()).DeleteVolumeRequest(oscgo.DeleteVolumeRequest{VolumeId: s.volumeId}).Execute()
	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting BSU volume: %s", err))
	}
}

func (s *StepCreateVolume) buildCreateVolumeInput(suregionName string, rootDevice *oscgo.BlockDeviceMappingImage) (*oscgo.CreateVolumeRequest, error) {
	if rootDevice == nil {
		return nil, fmt.Errorf("Couldn't find root device")
	}

	//FIX: Temporary fix
	gibSize := *rootDevice.Bsu.VolumeSize / (1024 * 1024 * 1024)
	createVolumeInput := &oscgo.CreateVolumeRequest{
		SubregionName: suregionName,
		Size:          &gibSize,
		SnapshotId:    rootDevice.Bsu.SnapshotId,
		VolumeType:    rootDevice.Bsu.VolumeType,
		Iops:          rootDevice.Bsu.Iops,
	}
	if int32(s.RootVolumeSize) > *rootDevice.Bsu.VolumeSize {
		*createVolumeInput.Size = int32(s.RootVolumeSize)
	}

	if s.RootVolumeType == "" || s.RootVolumeType == rootDevice.Bsu.GetVolumeType() {
		return createVolumeInput, nil
	}

	if s.RootVolumeType == "io1" {
		return nil, fmt.Errorf("Root volume type cannot be io1, because existing root volume type was %s", rootDevice.Bsu.GetVolumeType())
	}

	createVolumeInput.VolumeType = &s.RootVolumeType
	// non io1 cannot set iops
	*createVolumeInput.Iops = 0

	return createVolumeInput, nil
}
