package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/kubescape/go-logger"
	"github.com/kubescape/go-logger/helpers"
	"github.com/Aryaman6492/node-agent/pkg/config"
	"github.com/Aryaman6492/node-agent/pkg/storage"
	"github.com/Aryaman6492/storage/pkg/apis/softwarecomposition"
	"github.com/Aryaman6492/storage/pkg/apis/softwarecomposition/v1beta1"
	"github.com/Aryaman6492/storage/pkg/generated/clientset/versioned"
	"github.com/Aryaman6492/storage/pkg/generated/clientset/versioned/fake"
	spdxv1beta1 "github.com/Aryaman6492/storage/pkg/generated/clientset/versioned/typed/softwarecomposition/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	KubeConfig = "KUBECONFIG"
)

type Storage struct {
	StorageClient          spdxv1beta1.SpdxV1beta1Interface
	maxElapsedTime         time.Duration
	maxJsonPatchOperations int
	namespace              string
	multiplier             *int // used for testing to multiply the resources by this
}

var _ storage.StorageClient = (*Storage)(nil)

func CreateStorage(namespace string, maxElapsedTime time.Duration) (*Storage, error) {
	var cfg *rest.Config
	kubeconfig := os.Getenv(KubeConfig)
	// use the current context in kubeconfig
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		cfg, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create K8S Aggregated API Client with err: %v", err)
		}
	}
	// force GRPC
	cfg.AcceptContentTypes = "application/vnd.kubernetes.protobuf"
	cfg.ContentType = "application/vnd.kubernetes.protobuf"

	clientset, err := versioned.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create K8S Aggregated API Client with err: %v", err)
	}

	// wait for storage to be ready
	if err := backoff.RetryNotify(func() error {
		_, err := clientset.SpdxV1beta1().ApplicationProfiles("default").List(context.Background(), metav1.ListOptions{})
		return err
	}, backoff.WithMaxRetries(backoff.NewConstantBackOff(5*time.Second), 60), func(err error, d time.Duration) {
		logger.L().Info("waiting for storage to be ready", helpers.Error(err), helpers.String("retry in", d.String()))
	}); err != nil {
		return nil, fmt.Errorf("too many retries waiting for storage: %w", err)
	}

	return &Storage{
		StorageClient:          clientset.SpdxV1beta1(),
		maxElapsedTime:         maxElapsedTime,
		maxJsonPatchOperations: 9999,
		namespace:              namespace,
		multiplier:             getMultiplier(),
	}, nil
}

func CreateFakeStorage(namespace string) (*Storage, error) {
	return &Storage{
		StorageClient: fake.NewSimpleClientset().SpdxV1beta1(),
		namespace:     namespace,
	}, nil
}

func (sc Storage) CreateNetworkNeighbors(networkNeighbors *v1beta1.NetworkNeighbors, namespace string) error {
	sc.modifyNameP(&networkNeighbors.Name)
	defer sc.modifyNameP(&networkNeighbors.Name)

	_, err := sc.StorageClient.NetworkNeighborses(namespace).Create(context.Background(), networkNeighbors, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (sc Storage) GetNetworkNeighbors(namespace, name string) (*v1beta1.NetworkNeighbors, error) {
	nn, err := sc.StorageClient.NetworkNeighborses(namespace).Get(context.Background(), sc.modifyName(name), metav1.GetOptions{})
	if nn != nil {
		sc.revertNameP(&nn.Name)
	}
	return nn, err
}

func (sc Storage) PatchNetworkNeighborsIngressAndEgress(name, namespace string, networkNeighbors *v1beta1.NetworkNeighbors) error {
	sc.modifyNameP(&networkNeighbors.Name)
	defer sc.revertNameP(&networkNeighbors.Name)

	bytes, err := json.Marshal(networkNeighbors)
	if err != nil {
		return err
	}

	_, err = sc.StorageClient.NetworkNeighborses(namespace).Patch(context.Background(), sc.modifyName(name), types.StrategicMergePatchType, bytes, metav1.PatchOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (sc Storage) PatchNetworkNeighborsMatchLabels(_, namespace string, networkNeighbors *v1beta1.NetworkNeighbors) error {
	sc.modifyNameP(&networkNeighbors.Name)
	defer sc.revertNameP(&networkNeighbors.Name)
	_, err := sc.StorageClient.NetworkNeighborses(namespace).Update(context.Background(), networkNeighbors, metav1.UpdateOptions{})

	return err
}

func (sc Storage) CreateApplicationActivity(activity *v1beta1.ApplicationActivity, namespace string) error {
	sc.modifyNameP(&activity.Name)
	defer sc.revertNameP(&activity.Name)

	_, err := sc.StorageClient.ApplicationActivities(namespace).Create(context.Background(), activity, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (sc Storage) GetApplicationActivity(namespace, name string) (*v1beta1.ApplicationActivity, error) {

	aa, err := sc.StorageClient.ApplicationActivities(namespace).Get(context.Background(), sc.modifyName(name), metav1.GetOptions{})
	if aa != nil {
		sc.revertNameP(&aa.Name)
	}
	return aa, err
}

func (sc Storage) CreateSBOM(SBOM *v1beta1.SBOMSyft) (*v1beta1.SBOMSyft, error) {
	return sc.StorageClient.SBOMSyfts(sc.namespace).Create(context.Background(), SBOM, metav1.CreateOptions{})
}

func (sc Storage) GetSBOM(name string) (*v1beta1.SBOMSyft, error) {
	return sc.StorageClient.SBOMSyfts(sc.namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (sc Storage) GetSBOMMeta(name string) (*v1beta1.SBOMSyft, error) {
	return sc.StorageClient.SBOMSyfts(sc.namespace).Get(context.Background(), name, metav1.GetOptions{ResourceVersion: softwarecomposition.ResourceVersionMetadata})
}

func (sc Storage) ReplaceSBOM(SBOM *v1beta1.SBOMSyft) (*v1beta1.SBOMSyft, error) {
	return sc.StorageClient.SBOMSyfts(sc.namespace).Update(context.Background(), SBOM, metav1.UpdateOptions{})
}

func (sc Storage) IncrementImageUse(_ string) {
	// noop
}

func (sc Storage) DecrementImageUse(_ string) {
	// noop
}

func (sc Storage) modifyName(n string) string {
	if sc.multiplier != nil {
		return fmt.Sprintf("%s-%d", n, *sc.multiplier)
	}
	return n
}

func (sc Storage) modifyNameP(n *string) {
	if sc.multiplier != nil {
		*n = fmt.Sprintf("%s-%d", *n, *sc.multiplier)
	}
}

func (sc Storage) revertNameP(n *string) {
	if sc.multiplier != nil {
		*n = strings.TrimSuffix(*n, fmt.Sprintf("-%d", *sc.multiplier))
	}
}

func getMultiplier() *int {
	if m := os.Getenv("MULTIPLY"); m != "true" {
		return nil
	}
	podName := os.Getenv(config.PodNameEnvVar)
	s := strings.Split(podName, "-")
	if len(s) > 0 {
		if m, err := strconv.Atoi(s[len(s)-1]); err == nil {
			return &m
		}
	}
	return nil
}
