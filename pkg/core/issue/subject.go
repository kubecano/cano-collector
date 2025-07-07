package issue

import (
	"fmt"
	"strings"
)

// SubjectType represents the type of Kubernetes resource
type SubjectType int

const (
	SubjectTypeNone SubjectType = iota
	SubjectTypePod
	SubjectTypeDeployment
	SubjectTypeService
	SubjectTypeNode
	SubjectTypeNamespace
	SubjectTypeJob
	SubjectTypeCronJob
	SubjectTypeDaemonSet
	SubjectTypeStatefulSet
	SubjectTypeReplicaSet
	SubjectTypeIngress
	SubjectTypeConfigMap
	SubjectTypeSecret
	SubjectTypePersistentVolume
	SubjectTypePersistentVolumeClaim
	SubjectTypeCluster
	SubjectTypeHPA
)

// String returns the string representation of the subject type
func (st SubjectType) String() string {
	switch st {
	case SubjectTypeNone:
		return "NONE"
	case SubjectTypePod:
		return "POD"
	case SubjectTypeDeployment:
		return "DEPLOYMENT"
	case SubjectTypeService:
		return "SERVICE"
	case SubjectTypeNode:
		return "NODE"
	case SubjectTypeNamespace:
		return "NAMESPACE"
	case SubjectTypeJob:
		return "JOB"
	case SubjectTypeCronJob:
		return "CRONJOB"
	case SubjectTypeDaemonSet:
		return "DAEMONSET"
	case SubjectTypeStatefulSet:
		return "STATEFULSET"
	case SubjectTypeReplicaSet:
		return "REPLICASET"
	case SubjectTypeIngress:
		return "INGRESS"
	case SubjectTypeConfigMap:
		return "CONFIGMAP"
	case SubjectTypeSecret:
		return "SECRET"
	case SubjectTypePersistentVolume:
		return "PERSISTENTVOLUME"
	case SubjectTypePersistentVolumeClaim:
		return "PERSISTENTVOLUMECLAIM"
	case SubjectTypeCluster:
		return "CLUSTER"
	case SubjectTypeHPA:
		return "HPA"
	default:
		return "UNKNOWN"
	}
}

// FromString converts a string to SubjectType
func SubjectTypeFromString(s string) (SubjectType, error) {
	switch strings.ToUpper(s) {
	case "NONE":
		return SubjectTypeNone, nil
	case "POD":
		return SubjectTypePod, nil
	case "DEPLOYMENT":
		return SubjectTypeDeployment, nil
	case "SERVICE":
		return SubjectTypeService, nil
	case "NODE":
		return SubjectTypeNode, nil
	case "NAMESPACE":
		return SubjectTypeNamespace, nil
	case "JOB":
		return SubjectTypeJob, nil
	case "CRONJOB":
		return SubjectTypeCronJob, nil
	case "DAEMONSET":
		return SubjectTypeDaemonSet, nil
	case "STATEFULSET":
		return SubjectTypeStatefulSet, nil
	case "REPLICASET":
		return SubjectTypeReplicaSet, nil
	case "INGRESS":
		return SubjectTypeIngress, nil
	case "CONFIGMAP":
		return SubjectTypeConfigMap, nil
	case "SECRET":
		return SubjectTypeSecret, nil
	case "PERSISTENTVOLUME":
		return SubjectTypePersistentVolume, nil
	case "PERSISTENTVOLUMECLAIM":
		return SubjectTypePersistentVolumeClaim, nil
	case "CLUSTER":
		return SubjectTypeCluster, nil
	case "HPA":
		return SubjectTypeHPA, nil
	default:
		return SubjectTypeNone, fmt.Errorf("unknown subject type: %s", s)
	}
}

// Subject represents information about the Kubernetes resource related to the issue
type Subject struct {
	Name        string            `json:"name"`
	SubjectType SubjectType       `json:"subject_type"`
	Namespace   string            `json:"namespace,omitempty"`
	Node        string            `json:"node,omitempty"`
	Container   string            `json:"container,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// NewSubject creates a new Subject
func NewSubject(name string, subjectType SubjectType) *Subject {
	return &Subject{
		Name:        name,
		SubjectType: subjectType,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}
}

// String returns a string representation of the subject
func (s *Subject) String() string {
	if s.Namespace != "" {
		return fmt.Sprintf("%s/%s (%s)", s.Namespace, s.Name, s.SubjectType)
	}
	return fmt.Sprintf("%s (%s)", s.Name, s.SubjectType)
}

// FormatWithEmoji returns a formatted string representation of the subject with emoji
func (s *Subject) FormatWithEmoji() string {
	if s.Namespace != "" {
		return fmt.Sprintf("🎯 Subject: %s/%s (%s)",
			s.Namespace, s.Name, s.SubjectType.String())
	}
	return fmt.Sprintf("🎯 Subject: %s (%s)",
		s.Name, s.SubjectType.String())
}
