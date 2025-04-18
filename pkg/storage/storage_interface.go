package storage

import (
	"github.com/Aryaman6492/node-agent/pkg/utils"
	"github.com/Aryaman6492/storage/pkg/apis/softwarecomposition/v1beta1"
)

type StorageClient interface {
	CreateApplicationActivity(activity *v1beta1.ApplicationActivity, namespace string) error
	CreateApplicationProfile(profile *v1beta1.ApplicationProfile, namespace string) error
	PatchApplicationProfile(name, namespace string, operations []utils.PatchOperation, channel chan error) error
	GetApplicationProfile(namespace, name string) (*v1beta1.ApplicationProfile, error)
	CreateSBOM(SBOM *v1beta1.SBOMSyft) (*v1beta1.SBOMSyft, error)
	GetSBOM(name string) (*v1beta1.SBOMSyft, error)
	GetSBOMMeta(name string) (*v1beta1.SBOMSyft, error)
	ReplaceSBOM(SBOM *v1beta1.SBOMSyft) (*v1beta1.SBOMSyft, error)
	IncrementImageUse(imageID string)
	DecrementImageUse(imageID string)
	GetNetworkNeighbors(namespace, name string) (*v1beta1.NetworkNeighbors, error)
	CreateNetworkNeighbors(networkNeighbors *v1beta1.NetworkNeighbors, namespace string) error
	PatchNetworkNeighborsMatchLabels(name, namespace string, networkNeighbors *v1beta1.NetworkNeighbors) error
	PatchNetworkNeighborsIngressAndEgress(name, namespace string, networkNeighbors *v1beta1.NetworkNeighbors) error
	GetNetworkNeighborhood(namespace, name string) (*v1beta1.NetworkNeighborhood, error)
	CreateNetworkNeighborhood(neighborhood *v1beta1.NetworkNeighborhood, namespace string) error
	PatchNetworkNeighborhood(name, namespace string, operations []utils.PatchOperation, channel chan error) error
}
