package common

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"

	oscgo "github.com/outscale/osc-sdk-go/v2"
)

func TestBlockDevice_LaunchDevices(t *testing.T) {
	cases := []struct {
		Config *BlockDevice
		Result oscgo.BlockDeviceMappingVmCreation
	}{
		{
			Config: &BlockDevice{
				DeviceName:         "/dev/sdb",
				SnapshotId:         "snap-1234",
				VolumeType:         "standard",
				VolumeSize:         8,
				DeleteOnVmDeletion: true,
			},

			Result: oscgo.BlockDeviceMappingVmCreation{
				DeviceName: aws.String("/dev/sdb"),
				Bsu: &oscgo.BsuToCreate{
					SnapshotId:         aws.String("snap-1234"),
					VolumeType:         aws.String("standard"),
					VolumeSize:         aws.Int32(8),
					DeleteOnVmDeletion: aws.Bool(true),
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName: "/dev/sdb",
				VolumeSize: 8,
			},

			Result: oscgo.BlockDeviceMappingVmCreation{
				DeviceName: aws.String("/dev/sdb"),
				Bsu: &oscgo.BsuToCreate{
					VolumeSize:         aws.Int32(8),
					DeleteOnVmDeletion: aws.Bool(false),
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:         "/dev/sdb",
				VolumeType:         "io1",
				VolumeSize:         8,
				DeleteOnVmDeletion: true,
				IOPS:               1000,
			},

			Result: oscgo.BlockDeviceMappingVmCreation{
				DeviceName: aws.String("/dev/sdb"),
				Bsu: &oscgo.BsuToCreate{
					VolumeType:         aws.String("io1"),
					VolumeSize:         aws.Int32(8),
					DeleteOnVmDeletion: aws.Bool(true),
					Iops:               aws.Int32(1000),
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:         "/dev/sdb",
				VolumeType:         "gp2",
				VolumeSize:         8,
				DeleteOnVmDeletion: true,
			},

			Result: oscgo.BlockDeviceMappingVmCreation{
				DeviceName: aws.String("/dev/sdb"),
				Bsu: &oscgo.BsuToCreate{
					VolumeType:         aws.String("gp2"),
					VolumeSize:         aws.Int32(8),
					DeleteOnVmDeletion: aws.Bool(true),
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:         "/dev/sdb",
				VolumeType:         "gp2",
				VolumeSize:         8,
				DeleteOnVmDeletion: true,
			},

			Result: oscgo.BlockDeviceMappingVmCreation{
				DeviceName: aws.String("/dev/sdb"),
				Bsu: &oscgo.BsuToCreate{
					VolumeType:         aws.String("gp2"),
					VolumeSize:         aws.Int32(8),
					DeleteOnVmDeletion: aws.Bool(true),
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:         "/dev/sdb",
				VolumeType:         "standard",
				DeleteOnVmDeletion: true,
			},

			Result: oscgo.BlockDeviceMappingVmCreation{
				DeviceName: aws.String("/dev/sdb"),
				Bsu: &oscgo.BsuToCreate{
					VolumeType:         aws.String("standard"),
					DeleteOnVmDeletion: aws.Bool(true),
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:  "/dev/sdb",
				VirtualName: "ephemeral0",
			},

			Result: oscgo.BlockDeviceMappingVmCreation{
				DeviceName:        aws.String("/dev/sdb"),
				VirtualDeviceName: aws.String("ephemeral0"),
			},
		},
		{
			Config: &BlockDevice{
				DeviceName: "/dev/sdb",
				NoDevice:   true,
			},

			Result: oscgo.BlockDeviceMappingVmCreation{
				DeviceName: aws.String("/dev/sdb"),
				NoDevice:   aws.String(""),
			},
		},
	}

	for _, tc := range cases {

		launchBlockDevices := LaunchBlockDevices{
			LaunchMappings: []BlockDevice{*tc.Config},
		}

		expected := []oscgo.BlockDeviceMappingVmCreation{tc.Result}

		launchResults := launchBlockDevices.BuildOSCLaunchDevices()
		if !reflect.DeepEqual(expected, launchResults) {
			t.Fatalf("Bad block device, \nexpected: %#v\n\ngot: %#v",
				expected, launchResults)
		}
	}
}

func TestBlockDevice_OMI(t *testing.T) {
	cases := []struct {
		Config *BlockDevice
		Result oscgo.BlockDeviceMappingImage
	}{
		{
			Config: &BlockDevice{
				DeviceName:         "/dev/sdb",
				SnapshotId:         "snap-1234",
				VolumeType:         "standard",
				VolumeSize:         8,
				DeleteOnVmDeletion: true,
			},

			Result: oscgo.BlockDeviceMappingImage{
				DeviceName: aws.String("/dev/sdb"),
				Bsu: &oscgo.BsuToCreate{
					SnapshotId:         aws.String("snap-1234"),
					VolumeType:         aws.String("standard"),
					VolumeSize:         aws.Int32(8),
					DeleteOnVmDeletion: aws.Bool(true),
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:         "/dev/sdb",
				VolumeSize:         8,
				DeleteOnVmDeletion: true,
			},

			Result: oscgo.BlockDeviceMappingImage{
				DeviceName: aws.String("/dev/sdb"),
				Bsu: &oscgo.BsuToCreate{
					VolumeSize:         aws.Int32(8),
					DeleteOnVmDeletion: aws.Bool(true),
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:         "/dev/sdb",
				VolumeType:         "io1",
				VolumeSize:         8,
				DeleteOnVmDeletion: true,
				IOPS:               1000,
			},

			Result: oscgo.BlockDeviceMappingImage{
				DeviceName: aws.String("/dev/sdb"),
				Bsu: &oscgo.BsuToCreate{
					VolumeType:         aws.String("io1"),
					VolumeSize:         aws.Int32(8),
					DeleteOnVmDeletion: aws.Bool(true),
					Iops:               aws.Int32(1000),
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:         "/dev/sdb",
				VolumeType:         "gp2",
				VolumeSize:         8,
				DeleteOnVmDeletion: true,
			},

			Result: oscgo.BlockDeviceMappingImage{
				DeviceName: aws.String("/dev/sdb"),
				Bsu: &oscgo.BsuToCreate{
					VolumeType:         aws.String("gp2"),
					VolumeSize:         aws.Int32(8),
					DeleteOnVmDeletion: aws.Bool(true),
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:         "/dev/sdb",
				VolumeType:         "gp2",
				VolumeSize:         8,
				DeleteOnVmDeletion: true,
			},

			Result: oscgo.BlockDeviceMappingImage{
				DeviceName: aws.String("/dev/sdb"),
				Bsu: &oscgo.BsuToCreate{
					VolumeType:         aws.String("gp2"),
					VolumeSize:         aws.Int32(8),
					DeleteOnVmDeletion: aws.Bool(true),
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:         "/dev/sdb",
				VolumeType:         "standard",
				DeleteOnVmDeletion: true,
			},

			Result: oscgo.BlockDeviceMappingImage{
				DeviceName: aws.String("/dev/sdb"),
				Bsu: &oscgo.BsuToCreate{
					VolumeType:         aws.String("standard"),
					DeleteOnVmDeletion: aws.Bool(true),
				},
			},
		},
		{
			Config: &BlockDevice{
				DeviceName:  "/dev/sdb",
				VirtualName: "ephemeral0",
			},

			Result: oscgo.BlockDeviceMappingImage{
				DeviceName:        aws.String("/dev/sdb"),
				VirtualDeviceName: aws.String("ephemeral0"),
			},
		},
	}

	for i, tc := range cases {
		omiBlockDevices := OMIBlockDevices{
			OMIMappings: []BlockDevice{*tc.Config},
		}

		expected := []oscgo.BlockDeviceMappingImage{tc.Result}

		omiResults := omiBlockDevices.BuildOscOMIDevices()
		if !reflect.DeepEqual(expected, omiResults) {
			t.Fatalf("%d - Bad block device, \nexpected: %+#v\n\ngot: %+#v",
				i, expected, omiResults)
		}
	}
}
