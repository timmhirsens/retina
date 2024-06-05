//go:unit

package ciliumutil

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"

	v2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	ciliumv2 "github.com/cilium/cilium/pkg/k8s/client/clientset/versioned/typed/cilium.io/v2"
	"github.com/cilium/cilium/pkg/k8s/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

// ensure all interfaces are implemented
var _ ciliumv2.CiliumEndpointInterface = &MockEndpointClient{}

type MockEndpointClient struct {
	l               logrus.FieldLogger
	namespace       string
	ciliumEndpoints *MockResource[*v2.CiliumEndpoint]
	watchers        []watch.Interface
}

func NewMockEndpointClient(l logrus.FieldLogger, namespace string, ciliumEndpoints *MockResource[*v2.CiliumEndpoint]) *MockEndpointClient {
	return &MockEndpointClient{
		l:               l,
		namespace:       namespace,
		ciliumEndpoints: ciliumEndpoints,
		watchers:        make([]watch.Interface, 0),
	}
}

func (m *MockEndpointClient) Create(ctx context.Context, ciliumEndpoint *v2.CiliumEndpoint, opts v1.CreateOptions) (*v2.CiliumEndpoint, error) {
	m.l.Info("MockEndpointClient.Create() called")
	_, ok, err := m.ciliumEndpoints.GetByKey(resource.NewKey(ciliumEndpoint))
	if err != nil {
		return nil, err
	}
	if ok {
		return nil, ErrAlreadyExists
	}

	m.ciliumEndpoints.Upsert(ciliumEndpoint)
	return ciliumEndpoint, nil
}

func (m *MockEndpointClient) Update(ctx context.Context, ciliumEndpoint *v2.CiliumEndpoint, opts v1.UpdateOptions) (*v2.CiliumEndpoint, error) {
	m.l.Info("MockEndpointClient.Update() called")
	m.ciliumEndpoints.cache[resource.NewKey(ciliumEndpoint)] = ciliumEndpoint
	return ciliumEndpoint, nil
}

func (m *MockEndpointClient) UpdateStatus(ctx context.Context, ciliumEndpoint *v2.CiliumEndpoint, opts v1.UpdateOptions) (*v2.CiliumEndpoint, error) {
	m.l.Warn("MockEndpointClient.UpdateStatus() called but this returns nil because it's not implemented")
	return nil, ErrNotImplemented
}

func (m *MockEndpointClient) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	m.l.Info("MockEndpointClient.Delete() called")
	_, ok, err := m.ciliumEndpoints.GetByKey(resource.Key{Name: name, Namespace: m.namespace})
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotFound{}
	}
	m.ciliumEndpoints.Delete(resource.Key{Name: name, Namespace: m.namespace})
	return nil
}

func (m *MockEndpointClient) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	m.l.Warn("MockEndpointClient.DeleteCollection() called but this is not implemented")
	return ErrNotImplemented
}

func (m *MockEndpointClient) Get(ctx context.Context, name string, opts v1.GetOptions) (*v2.CiliumEndpoint, error) {
	m.l.Info("MockEndpointClient.Get() called")
	item, _, err := m.ciliumEndpoints.GetByKey(resource.Key{Name: name, Namespace: m.namespace})
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (m *MockEndpointClient) List(ctx context.Context, opts v1.ListOptions) (*v2.CiliumEndpointList, error) {
	m.l.Info("MockEndpointClient.List() called")

	items := make([]v2.CiliumEndpoint, len(m.ciliumEndpoints.cache))
	for _, cep := range m.ciliumEndpoints.cache {
		items = append(items, *cep)
	}

	return &v2.CiliumEndpointList{Items: items}, nil
}

func (m *MockEndpointClient) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	m.l.Warn("MockEndpointClient.Watch() called but this returns a fake watch because it's not implemented")

	// not sure if watching is important for us
	w := watch.NewFake()
	m.watchers = append(m.watchers, w)
	return w, nil
}

func (m *MockEndpointClient) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v2.CiliumEndpoint, err error) {
	key := resource.Key{Name: name, Namespace: m.namespace}
	cep, ok, err := m.ciliumEndpoints.GetByKey(key)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, ErrNotFound{}
	}

	var replaceCEPStatus []JSONPatch
	err = json.Unmarshal(data, &replaceCEPStatus)
	if err != nil {
		return nil, err
	}

	cep.Status = replaceCEPStatus[0].Value
	m.ciliumEndpoints.Upsert(cep)
	cep, ok, err = m.ciliumEndpoints.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotFound{}
	}

	return cep, nil
}

type JSONPatch struct {
	OP    string            `json:"op,omitempty"`
	Path  string            `json:"path,omitempty"`
	Value v2.EndpointStatus `json:"value"`
}
