package fetch

import (
	"context"

	"github.com/operator-framework/catalogd/api/core/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
)

type CatalogFilterFunc func(catalog *v1alpha1.ClusterCatalog) bool
type CatalogFetcher interface {
	FetchCatalogs(ctx context.Context, filters ...CatalogFilterFunc) ([]v1alpha1.ClusterCatalog, error)
}

func New(client dynamic.Interface) CatalogFetcher {
	return &instance{
		client: client,
	}
}

type instance struct {
	client dynamic.Interface
}

func (c *instance) FetchCatalogs(ctx context.Context, filters ...CatalogFilterFunc) ([]v1alpha1.ClusterCatalog, error) {
	catalogList := &v1alpha1.ClusterCatalogList{}
	unstructCatalogs, err := c.client.Resource(v1alpha1.GroupVersion.WithResource("clustercatalogs")).List(ctx, v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructCatalogs.UnstructuredContent(), catalogList)
	if err != nil {
		return nil, err
	}

	catalogs := []v1alpha1.ClusterCatalog{}
	for _, catalog := range catalogList.Items {
		filteredOut := false
		for _, filter := range filters {
			if !filter(&catalog) {
				filteredOut = true
			}
		}

		if filteredOut {
			continue
		}

		catalogs = append(catalogs, catalog)
	}

	return catalogs, nil
}

func WithNameFilter(name string) CatalogFilterFunc {
	return func(catalog *v1alpha1.ClusterCatalog) bool {
		if name == "" {
			return true
		}
		return catalog.Name == name
	}
}

func WithUnpackedFilter() CatalogFilterFunc {
	return func(catalog *v1alpha1.ClusterCatalog) bool {
		return meta.IsStatusConditionTrue(catalog.Status.Conditions, v1alpha1.TypeUnpacked)
	}
}
