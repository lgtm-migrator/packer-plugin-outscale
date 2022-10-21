package bsu

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	oscgo "github.com/outscale/osc-sdk-go/v2"
	osccommon "github.com/outscale/packer-plugin-outscale/builder/osc/common"
)

type stepCreateOMI struct {
	image     *oscgo.Image
	RawRegion string
}

func (s *stepCreateOMI) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	oscconn := state.Get("osc").(*osccommon.OscClient)
	vm := state.Get("vm").(oscgo.Vm)
	ui := state.Get("ui").(packersdk.Ui)

	// Create the image
	omiName := config.OMIName

	ui.Say(fmt.Sprintf("Creating OMI %s from vm %s", omiName, vm.GetVmId()))
	blockDeviceMapping := config.BlockDevices.BuildOscOMIDevices()
	createOpts := oscgo.CreateImageRequest{
		VmId:                vm.VmId,
		ImageName:           &omiName,
		BlockDeviceMappings: &blockDeviceMapping,
	}
	if config.OMIDescription != "" {
		createOpts.Description = &config.OMIDescription
	}

	resp, _, err := oscconn.Api.ImageApi.CreateImage(oscconn.Auth).CreateImageRequest(createOpts).Execute()
	if err != nil || resp.GetImage().ImageId == nil {
		err := fmt.Errorf("Error creating OMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	image := resp.GetImage()

	// Set the OMI ID in the state
	ui.Message(fmt.Sprintf("OMI: %s", image.GetImageId()))
	omis := make(map[string]string)
	omis[s.RawRegion] = image.GetImageId()
	state.Put("omis", omis)

	// Wait for the image to become ready
	ui.Say("Waiting for OMI to become ready...")
	if err := osccommon.WaitUntilOscImageAvailable(oscconn, *image.ImageId); err != nil {
		log.Printf("Error waiting for OMI: %s", err)
		req := oscgo.ReadImagesRequest{
			Filters: &oscgo.FiltersImage{ImageIds: &[]string{image.GetImageId()}},
		}
		imagesResp, _, err := oscconn.Api.ImageApi.ReadImages(oscconn.Auth).ReadImagesRequest(req).Execute()
		if err != nil {
			log.Printf("Unable to determine reason waiting for OMI failed: %s", err)
			err = fmt.Errorf("Unknown error waiting for OMI")
		} else {
			stateReason := imagesResp.GetImages()[0].GetStateComment()
			err = fmt.Errorf("Error waiting for OMI. Reason: %s", *stateReason.StateMessage)
		}

		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	req := oscgo.ReadImagesRequest{
		Filters: &oscgo.FiltersImage{ImageIds: &[]string{image.GetImageId()}},
	}
	imagesResp, _, err := oscconn.Api.ImageApi.ReadImages(oscconn.Auth).ReadImagesRequest(req).Execute()
	if err != nil {
		err := fmt.Errorf("Error searching for OMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	s.image = &imagesResp.GetImages()[0]

	snapshots := make(map[string][]string)
	blockMapping := imagesResp.GetImages()[0].BlockDeviceMappings
	for _, blockDeviceMapping := range *blockMapping {
		if blockDeviceMapping.Bsu.SnapshotId != nil {
			snapshots[s.RawRegion] = append(snapshots[s.RawRegion], *blockDeviceMapping.Bsu.SnapshotId)
		}
	}
	state.Put("snapshots", snapshots)

	return multistep.ActionContinue
}

func (s *stepCreateOMI) Cleanup(state multistep.StateBag) {
	if s.image == nil {
		return
	}

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	oscconn := state.Get("osc").(*osccommon.OscClient)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Deregistering the OMI because cancellation or error...")
	DeleteOpts := oscgo.DeleteImageRequest{ImageId: s.image.GetImageId()}
	_, _, err := oscconn.Api.ImageApi.DeleteImage(oscconn.Auth).DeleteImageRequest(DeleteOpts).Execute()
	if err != nil {
		ui.Error(fmt.Sprintf("Error Deleting OMI, may still be around: %s", err))
		return
	}
}
