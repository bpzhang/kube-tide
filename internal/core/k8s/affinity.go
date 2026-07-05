package k8s

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LabelSelector 定义标签选择器
type LabelSelector struct {
	MatchLabels      map[string]string            `json:"matchLabels,omitempty"`
	MatchExpressions []LabelSelectorRequirement   `json:"matchExpressions,omitempty"`
}

// LabelSelectorRequirement 定义标签选择器表达式
type LabelSelectorRequirement struct {
	Key      string   `json:"key"`
	Operator string   `json:"operator"`
	Values   []string `json:"values,omitempty"`
}

func convertNodeSelectorRequirements(reqs []NodeSelectorRequirement) []corev1.NodeSelectorRequirement {
	if len(reqs) == 0 {
		return nil
	}
	result := make([]corev1.NodeSelectorRequirement, 0, len(reqs))
	for _, req := range reqs {
		result = append(result, corev1.NodeSelectorRequirement{
			Key:      req.Key,
			Operator: corev1.NodeSelectorOperator(req.Operator),
			Values:   req.Values,
		})
	}
	return result
}

func convertNodeSelectorTerms(terms []NodeSelectorTerm) []corev1.NodeSelectorTerm {
	if len(terms) == 0 {
		return nil
	}
	result := make([]corev1.NodeSelectorTerm, 0, len(terms))
	for _, term := range terms {
		result = append(result, corev1.NodeSelectorTerm{
			MatchExpressions: convertNodeSelectorRequirements(term.MatchExpressions),
			MatchFields:      convertNodeSelectorRequirements(term.MatchFields),
		})
	}
	return result
}

func convertLabelSelectorToK8s(selector *LabelSelector) *metav1.LabelSelector {
	if selector == nil {
		return nil
	}
	if len(selector.MatchLabels) == 0 && len(selector.MatchExpressions) == 0 {
		return nil
	}
	k8sSelector := &metav1.LabelSelector{
		MatchLabels: selector.MatchLabels,
	}
	if len(selector.MatchExpressions) > 0 {
		exprs := make([]metav1.LabelSelectorRequirement, 0, len(selector.MatchExpressions))
		for _, expr := range selector.MatchExpressions {
			exprs = append(exprs, metav1.LabelSelectorRequirement{
				Key:      expr.Key,
				Operator: metav1.LabelSelectorOperator(expr.Operator),
				Values:   expr.Values,
			})
		}
		k8sSelector.MatchExpressions = exprs
	}
	return k8sSelector
}

func convertPodAffinityTermToK8s(term PodAffinityTerm) corev1.PodAffinityTerm {
	return corev1.PodAffinityTerm{
		LabelSelector: convertLabelSelectorToK8s(term.LabelSelector),
		Namespaces:    term.Namespaces,
		TopologyKey:   term.TopologyKey,
	}
}

func convertPodAffinityTermsToK8s(terms []PodAffinityTerm) []corev1.PodAffinityTerm {
	if len(terms) == 0 {
		return nil
	}
	result := make([]corev1.PodAffinityTerm, 0, len(terms))
	for _, term := range terms {
		if term.TopologyKey == "" {
			continue
		}
		result = append(result, convertPodAffinityTermToK8s(term))
	}
	return result
}

func convertWeightedPodAffinityTermsToK8s(terms []WeightedPodAffinityTerm) []corev1.WeightedPodAffinityTerm {
	if len(terms) == 0 {
		return nil
	}
	result := make([]corev1.WeightedPodAffinityTerm, 0, len(terms))
	for _, term := range terms {
		if term.PodAffinityTerm.TopologyKey == "" {
			continue
		}
		result = append(result, corev1.WeightedPodAffinityTerm{
			Weight:          term.Weight,
			PodAffinityTerm: convertPodAffinityTermToK8s(term.PodAffinityTerm),
		})
	}
	return result
}

func buildK8sAffinity(affinity *Affinity) *corev1.Affinity {
	if affinity == nil {
		return nil
	}

	k8sAffinity := &corev1.Affinity{}

	if affinity.NodeAffinity != nil {
		k8sNodeAffinity := &corev1.NodeAffinity{}
		if affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
			terms := convertNodeSelectorTerms(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms)
			if len(terms) > 0 {
				k8sNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{
					NodeSelectorTerms: terms,
				}
			}
		}
		if len(affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
			preferred := make([]corev1.PreferredSchedulingTerm, 0, len(affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution))
			for _, term := range affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
				preferred = append(preferred, corev1.PreferredSchedulingTerm{
					Weight: term.Weight,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: convertNodeSelectorRequirements(term.Preference.MatchExpressions),
						MatchFields:      convertNodeSelectorRequirements(term.Preference.MatchFields),
					},
				})
			}
			k8sNodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = preferred
		}
		if k8sNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil ||
			len(k8sNodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
			k8sAffinity.NodeAffinity = k8sNodeAffinity
		}
	}

	if affinity.PodAffinity != nil {
		k8sPodAffinity := &corev1.PodAffinity{}
		if required := convertPodAffinityTermsToK8s(affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution); len(required) > 0 {
			k8sPodAffinity.RequiredDuringSchedulingIgnoredDuringExecution = required
		}
		if preferred := convertWeightedPodAffinityTermsToK8s(affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution); len(preferred) > 0 {
			k8sPodAffinity.PreferredDuringSchedulingIgnoredDuringExecution = preferred
		}
		if k8sPodAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil ||
			len(k8sPodAffinity.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
			k8sAffinity.PodAffinity = k8sPodAffinity
		}
	}

	if affinity.PodAntiAffinity != nil {
		k8sPodAntiAffinity := &corev1.PodAntiAffinity{}
		if required := convertPodAffinityTermsToK8s(affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution); len(required) > 0 {
			k8sPodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = required
		}
		if preferred := convertWeightedPodAffinityTermsToK8s(affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution); len(preferred) > 0 {
			k8sPodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = preferred
		}
		if k8sPodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil ||
			len(k8sPodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
			k8sAffinity.PodAntiAffinity = k8sPodAntiAffinity
		}
	}

	if k8sAffinity.NodeAffinity == nil && k8sAffinity.PodAffinity == nil && k8sAffinity.PodAntiAffinity == nil {
		return nil
	}
	return k8sAffinity
}

func convertK8sNodeSelectorRequirements(reqs []corev1.NodeSelectorRequirement) []NodeSelectorRequirement {
	if len(reqs) == 0 {
		return nil
	}
	result := make([]NodeSelectorRequirement, 0, len(reqs))
	for _, req := range reqs {
		result = append(result, NodeSelectorRequirement{
			Key:      req.Key,
			Operator: string(req.Operator),
			Values:   req.Values,
		})
	}
	return result
}

func convertK8sLabelSelector(selector *metav1.LabelSelector) *LabelSelector {
	if selector == nil {
		return nil
	}
	if len(selector.MatchLabels) == 0 && len(selector.MatchExpressions) == 0 {
		return nil
	}
	result := &LabelSelector{
		MatchLabels: selector.MatchLabels,
	}
	if len(selector.MatchExpressions) > 0 {
		exprs := make([]LabelSelectorRequirement, 0, len(selector.MatchExpressions))
		for _, expr := range selector.MatchExpressions {
			exprs = append(exprs, LabelSelectorRequirement{
				Key:      expr.Key,
				Operator: string(expr.Operator),
				Values:   expr.Values,
			})
		}
		result.MatchExpressions = exprs
	}
	return result
}

func convertK8sPodAffinityTerm(term corev1.PodAffinityTerm) PodAffinityTerm {
	return PodAffinityTerm{
		LabelSelector: convertK8sLabelSelector(term.LabelSelector),
		Namespaces:    term.Namespaces,
		TopologyKey:   term.TopologyKey,
	}
}

func convertK8sPodAffinityTerms(terms []corev1.PodAffinityTerm) []PodAffinityTerm {
	if len(terms) == 0 {
		return nil
	}
	result := make([]PodAffinityTerm, 0, len(terms))
	for _, term := range terms {
		result = append(result, convertK8sPodAffinityTerm(term))
	}
	return result
}

func convertK8sWeightedPodAffinityTerms(terms []corev1.WeightedPodAffinityTerm) []WeightedPodAffinityTerm {
	if len(terms) == 0 {
		return nil
	}
	result := make([]WeightedPodAffinityTerm, 0, len(terms))
	for _, term := range terms {
		result = append(result, WeightedPodAffinityTerm{
			Weight:          term.Weight,
			PodAffinityTerm: convertK8sPodAffinityTerm(term.PodAffinityTerm),
		})
	}
	return result
}

func convertK8sAffinity(affinity *corev1.Affinity) *Affinity {
	if affinity == nil {
		return nil
	}

	result := &Affinity{}

	if affinity.NodeAffinity != nil {
		nodeAffinity := &NodeAffinity{}
		if affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
			terms := make([]NodeSelectorTerm, 0, len(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms))
			for _, term := range affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
				terms = append(terms, NodeSelectorTerm{
					MatchExpressions: convertK8sNodeSelectorRequirements(term.MatchExpressions),
					MatchFields:      convertK8sNodeSelectorRequirements(term.MatchFields),
				})
			}
			nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &NodeSelector{
				NodeSelectorTerms: terms,
			}
		}
		if len(affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
			preferred := make([]PreferredSchedulingTerm, 0, len(affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution))
			for _, term := range affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
				preferred = append(preferred, PreferredSchedulingTerm{
					Weight: term.Weight,
					Preference: NodeSelectorTerm{
						MatchExpressions: convertK8sNodeSelectorRequirements(term.Preference.MatchExpressions),
						MatchFields:      convertK8sNodeSelectorRequirements(term.Preference.MatchFields),
					},
				})
			}
			nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = preferred
		}
		if nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil ||
			len(nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
			result.NodeAffinity = nodeAffinity
		}
	}

	if affinity.PodAffinity != nil {
		podAffinity := &PodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution:  convertK8sPodAffinityTerms(affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution),
			PreferredDuringSchedulingIgnoredDuringExecution: convertK8sWeightedPodAffinityTerms(affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution),
		}
		if len(podAffinity.RequiredDuringSchedulingIgnoredDuringExecution) > 0 ||
			len(podAffinity.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
			result.PodAffinity = podAffinity
		}
	}

	if affinity.PodAntiAffinity != nil {
		podAntiAffinity := &PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution:  convertK8sPodAffinityTerms(affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution),
			PreferredDuringSchedulingIgnoredDuringExecution: convertK8sWeightedPodAffinityTerms(affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution),
		}
		if len(podAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) > 0 ||
			len(podAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
			result.PodAntiAffinity = podAntiAffinity
		}
	}

	if result.NodeAffinity == nil && result.PodAffinity == nil && result.PodAntiAffinity == nil {
		return nil
	}
	return result
}

func convertK8sTolerations(tolerations []corev1.Toleration) []Toleration {
	if len(tolerations) == 0 {
		return nil
	}
	result := make([]Toleration, 0, len(tolerations))
	for _, toleration := range tolerations {
		result = append(result, Toleration{
			Key:               toleration.Key,
			Operator:          string(toleration.Operator),
			Value:             toleration.Value,
			Effect:            string(toleration.Effect),
			TolerationSeconds: toleration.TolerationSeconds,
		})
	}
	return result
}

func convertTolerationsToK8s(tolerations []Toleration) []corev1.Toleration {
	if len(tolerations) == 0 {
		return nil
	}
	result := make([]corev1.Toleration, 0, len(tolerations))
	for _, toleration := range tolerations {
		result = append(result, corev1.Toleration{
			Key:               toleration.Key,
			Operator:          corev1.TolerationOperator(toleration.Operator),
			Value:             toleration.Value,
			Effect:            corev1.TaintEffect(toleration.Effect),
			TolerationSeconds: toleration.TolerationSeconds,
		})
	}
	return result
}
