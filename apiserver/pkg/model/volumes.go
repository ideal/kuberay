package model

import (
	api "github.com/ray-project/kuberay/proto/go_client"
	v1 "k8s.io/api/core/v1"
)

func PopulateVolumes(podTemplate *v1.PodTemplateSpec) []*api.Volume {
	if len(podTemplate.Spec.Volumes) == 0 {
		return nil
	}
	var volumes []*api.Volume
	for _, vol := range podTemplate.Spec.Volumes {
		mount := GetVolumeMount(podTemplate, vol.Name)
		if vol.VolumeSource.HostPath != nil {
			// Host Path
			volumes = append(volumes, &api.Volume{
				Name:                 vol.Name,
				MountPath:            mount.MountPath,
				Source:               vol.VolumeSource.HostPath.Path,
				MountPropagationMode: GetVolumeMountPropagation(mount),
				VolumeType:           api.Volume_VolumeType(api.Volume_HOSTTOCONTAINER),
				HostPathType:         GetVolumeHostPathType(&vol),
			})
			continue

		}
		if vol.VolumeSource.PersistentVolumeClaim != nil {
			// PVC
			volumes = append(volumes, &api.Volume{
				Name:                 vol.Name,
				MountPath:            mount.MountPath,
				MountPropagationMode: GetVolumeMountPropagation(mount),
				VolumeType:           api.Volume_PERSISTENT_VOLUME_CLAIM,
				ReadOnly:             vol.VolumeSource.PersistentVolumeClaim.ReadOnly,
			})
			continue
		}
		if vol.VolumeSource.Ephemeral != nil {
			// Ephimeral
			request := vol.VolumeSource.Ephemeral.VolumeClaimTemplate.Spec.Resources.Requests[v1.ResourceStorage]
			volume := api.Volume{
				Name:                 vol.Name,
				MountPath:            mount.MountPath,
				MountPropagationMode: GetVolumeMountPropagation(mount),
				VolumeType:           api.Volume_EPHEMERAL,
				AccessMode:           GetAccessMode(&vol),
				Storage:              request.String(),
			}
			if vol.VolumeSource.Ephemeral.VolumeClaimTemplate.Spec.StorageClassName != nil {
				volume.StorageClassName = *vol.VolumeSource.Ephemeral.VolumeClaimTemplate.Spec.StorageClassName
			}
			volumes = append(volumes, &volume)
			continue
		}
	}
	return volumes
}

func GetVolumeMount(podTemplate *v1.PodTemplateSpec, vol string) *v1.VolumeMount {
	for _, container := range podTemplate.Spec.Containers {
		for _, mount := range container.VolumeMounts {
			if mount.Name == vol {
				return &mount
			}
		}
	}
	return nil
}

func GetVolumeMountPropagation(mount *v1.VolumeMount) api.Volume_MountPropagationMode {
	if mount.MountPropagation == nil {
		return api.Volume_NONE
	}
	if *mount.MountPropagation == v1.MountPropagationHostToContainer {
		return api.Volume_HOSTTOCONTAINER
	}
	if *mount.MountPropagation == v1.MountPropagationBidirectional {
		return api.Volume_BIDIRECTIONAL
	}
	return api.Volume_NONE
}

func GetVolumeHostPathType(vol *v1.Volume) api.Volume_HostPathType {
	if *vol.VolumeSource.HostPath.Type == v1.HostPathFile {
		return api.Volume_FILE
	}
	return api.Volume_DIRECTORY
}

func GetAccessMode(vol *v1.Volume) api.Volume_AccessMode {
	modes := vol.VolumeSource.Ephemeral.VolumeClaimTemplate.Spec.AccessModes
	if len(modes) == 0 {
		return api.Volume_RWO
	}
	if modes[0] == v1.ReadOnlyMany {
		return api.Volume_ROX
	}
	if modes[0] == v1.ReadWriteMany {
		return api.Volume_RWX
	}
	return api.Volume_RWO
}
