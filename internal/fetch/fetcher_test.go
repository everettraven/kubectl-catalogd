package fetch

import (
	"context"
	"testing"

	"github.com/operator-framework/catalogd/api/core/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
)

func TestFetcher(t *testing.T) {
	var tests = []struct {
		name             string
		fetcher          CatalogFetcher
		filters          []CatalogFilterFunc
		expectedCatalogs []v1alpha1.ClusterCatalog
	}{
		{
			name: "no catalogs exist, no catalogs returned",
			fetcher: func() CatalogFetcher {
				scheme := runtime.NewScheme()
				err := v1alpha1.AddToScheme(scheme)
				require.NoError(t, err)

				dc := fake.NewSimpleDynamicClient(scheme)
				return New(dc)
			}(),
			expectedCatalogs: []v1alpha1.ClusterCatalog{},
		},
		{
			name: "catalogs exist, no filters, all catalogs returned",
			fetcher: func() CatalogFetcher {
				scheme := runtime.NewScheme()
				err := v1alpha1.AddToScheme(scheme)
				require.NoError(t, err)

				dc := fake.NewSimpleDynamicClient(scheme, &v1alpha1.ClusterCatalog{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-catalog",
					},
				})
				return New(dc)
			}(),
			expectedCatalogs: []v1alpha1.ClusterCatalog{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterCatalog",
						APIVersion: v1alpha1.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-catalog",
					},
				},
			},
		},
		{
			name: "catalogs exist, name filter, only matching catalogs returned",
			fetcher: func() CatalogFetcher {
				scheme := runtime.NewScheme()
				err := v1alpha1.AddToScheme(scheme)
				require.NoError(t, err)

				dc := fake.NewSimpleDynamicClient(scheme, &v1alpha1.ClusterCatalog{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-catalog",
					},
				}, &v1alpha1.ClusterCatalog{
					ObjectMeta: metav1.ObjectMeta{
						Name: "another-catalog",
					},
				})
				return New(dc)
			}(),
			expectedCatalogs: []v1alpha1.ClusterCatalog{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterCatalog",
						APIVersion: v1alpha1.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-catalog",
					},
				},
			},
			filters: []CatalogFilterFunc{
				WithNameFilter("test-catalog"),
			},
		},
		{
			name: "catalogs exist, unpacked filter, only matching catalogs returned",
			fetcher: func() CatalogFetcher {
				scheme := runtime.NewScheme()
				err := v1alpha1.AddToScheme(scheme)
				require.NoError(t, err)

				dc := fake.NewSimpleDynamicClient(scheme, &v1alpha1.ClusterCatalog{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-catalog",
					},
					Status: v1alpha1.ClusterCatalogStatus{
						Conditions: []metav1.Condition{
							{
								Type:   v1alpha1.TypeUnpacked,
								Status: metav1.ConditionTrue,
							},
						},
					},
				}, &v1alpha1.ClusterCatalog{
					ObjectMeta: metav1.ObjectMeta{
						Name: "another-catalog",
					},
				})
				return New(dc)
			}(),
			expectedCatalogs: []v1alpha1.ClusterCatalog{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterCatalog",
						APIVersion: v1alpha1.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-catalog",
					},
					Status: v1alpha1.ClusterCatalogStatus{
						Conditions: []metav1.Condition{
							{
								Type:   v1alpha1.TypeUnpacked,
								Status: metav1.ConditionTrue,
							},
						},
					},
				},
			},
			filters: []CatalogFilterFunc{
				WithUnpackedFilter(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			catalogs, err := tt.fetcher.FetchCatalogs(context.Background(), tt.filters...)
			require.NoError(t, err)
			require.Equal(t, tt.expectedCatalogs, catalogs)
		})
	}
}
