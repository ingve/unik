package xen

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"os"
)

func (p *XenProvider) DeleteInstance(id string, force bool) error {
	instance, err := p.GetInstance(id)
	if err != nil {
		return errors.New("retrieving instance "+id, err)
	}

	if err := p.client.DestroyVm(instance.Name); err != nil {
		return errors.New("xen api destroy call", err)
	}

	if err := os.RemoveAll(getInstanceDir(instance.Name)); err != nil {
		return errors.New("deleting instance dir", err)
	}

	volumesToDetach := []*types.Volume{}
	volumes, err := p.ListVolumes()
	if err != nil {
		return errors.New("getting volume list", err)
	}
	for _, volume := range volumes {
		if volume.Attachment == instance.Id {
			volumesToDetach = append(volumesToDetach, volume)
		}
	}
	if err := p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
		delete(instances, instance.Id)
		return nil
	}); err != nil {
		return errors.New("modifying image map in state", err)
	}
	for _, volume := range volumesToDetach {
		if err := p.state.ModifyVolumes(func(volumes map[string]*types.Volume) error {
			volume, ok := volumes[volume.Id]
			if !ok {
				return errors.New("no record of "+volume.Id+" in the state", nil)
			}
			volume.Attachment = ""
			return nil
		}); err != nil {
			return errors.New("modifying volume map in state", err)
		}
	}
	return nil
}
