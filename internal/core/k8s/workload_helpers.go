package k8s

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func buildContainerFromSpec(spec ContainerSpec) corev1.Container {
	c := corev1.Container{
		Name:            spec.Name,
		Image:           spec.Image,
		Command:         spec.Command,
		Args:            spec.Args,
		WorkingDir:      spec.WorkingDir,
		ImagePullPolicy: corev1.PullPolicy(spec.ImagePullPolicy),
	}
	if len(spec.Ports) > 0 {
		c.Ports = make([]corev1.ContainerPort, 0, len(spec.Ports))
		for _, p := range spec.Ports {
			protocol := corev1.ProtocolTCP
			if p.Protocol != "" {
				protocol = corev1.Protocol(p.Protocol)
			}
			c.Ports = append(c.Ports, corev1.ContainerPort{
				Name:          p.Name,
				ContainerPort: p.ContainerPort,
				HostPort:      p.HostPort,
				Protocol:      protocol,
			})
		}
	}
	if len(spec.Env) > 0 {
		c.Env = make([]corev1.EnvVar, 0, len(spec.Env))
		for _, env := range spec.Env {
			ev := corev1.EnvVar{Name: env.Name, Value: env.Value}
			if env.ValueFrom != nil {
				ev.ValueFrom = &corev1.EnvVarSource{}
				if env.ValueFrom.ConfigMapKeyRef != nil {
					ev.ValueFrom.ConfigMapKeyRef = &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: env.ValueFrom.ConfigMapKeyRef.Name},
						Key:                  env.ValueFrom.ConfigMapKeyRef.Key,
					}
				}
				if env.ValueFrom.SecretKeyRef != nil {
					ev.ValueFrom.SecretKeyRef = &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: env.ValueFrom.SecretKeyRef.Name},
						Key:                  env.ValueFrom.SecretKeyRef.Key,
					}
				}
			}
			c.Env = append(c.Env, ev)
		}
	}
	if spec.Resources.Limits != nil || spec.Resources.Requests != nil {
		c.Resources = buildResourceRequirements(spec.Resources)
	}
	return c
}

func buildResourceRequirements(res ResourceRequirements) corev1.ResourceRequirements {
	result := corev1.ResourceRequirements{}
	if len(res.Limits) > 0 {
		result.Limits = corev1.ResourceList{}
		for k, v := range res.Limits {
			result.Limits[corev1.ResourceName(k)] = resource.MustParse(v)
		}
	}
	if len(res.Requests) > 0 {
		result.Requests = corev1.ResourceList{}
		for k, v := range res.Requests {
			result.Requests[corev1.ResourceName(k)] = resource.MustParse(v)
		}
	}
	return result
}

func applyResourceUpdates(containers *[]corev1.Container, resources map[string]ResourceRequirements) {
	for i, c := range *containers {
		if res, ok := resources[c.Name]; ok {
			(*containers)[i].Resources = buildResourceRequirements(res)
		}
	}
}
