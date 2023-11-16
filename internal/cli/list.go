package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/everettraven/kubectl-catalogd/internal/fetch"
	"github.com/everettraven/kubectl-catalogd/internal/stream"
	"github.com/everettraven/kubectl-catalogd/internal/styles"
	"github.com/operator-framework/operator-registry/alpha/declcfg"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

var listCmd = cobra.Command{
	Use:   "list [flags]",
	Short: "Lists catalog objects",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := ctrl.GetConfigOrDie()
		dynamicClient, err := dynamic.NewForConfig(cfg)
		if err != nil {
			return err
		}
		kubeClient, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			return err
		}

		fetcher := fetch.New(dynamicClient)
		streamer := stream.New(kubeClient.CoreV1())

		return list(fetcher, streamer, listCfg)
	},
}

type lister struct {
	schema      string
	pkg         string
	name        string
	catalogName string
}

var listCfg = lister{
	schema:      "",
	pkg:         "",
	name:        "",
	catalogName: "",
}

func init() {
	listCmd.Flags().StringVar(&listCfg.schema, "schema", "", "specify the FBC object schema that should be used to filter the resulting output")
	listCmd.Flags().StringVar(&listCfg.pkg, "package", "", "specify the FBC object package that should be used to filter the resulting output")
	listCmd.Flags().StringVar(&listCfg.name, "name", "", "specify the FBC object name that should be used to filter the resulting output")
	listCmd.Flags().StringVar(&listCfg.catalogName, "catalog", "", "specify the catalog that should be used. By default it will fetch from all catalogs")
}

func list(fetcher fetch.CatalogFetcher, streamer stream.CatalogContentStreamer, listCfg lister) error {
	ctx := context.Background()
	catalogs, err := fetcher.FetchCatalogs(ctx, fetch.WithNameFilter(listCfg.catalogName), fetch.WithUnpackedFilter())
	if err != nil {
		return err
	}

	for _, catalog := range catalogs {
		rc, err := streamer.StreamCatalogContents(ctx, catalog)
		if err != nil {
			return fmt.Errorf("streaming FBC for catalog %q: %w", catalog.Name, err)
		}
		err = declcfg.WalkMetasReader(rc, func(meta *declcfg.Meta, err error) error {
			if err != nil {
				return err
			}

			if listCfg.schema != "" && meta.Schema != listCfg.schema {
				return nil
			}

			if listCfg.pkg != "" && meta.Package != listCfg.pkg {
				return nil
			}

			if listCfg.name != "" && meta.Name != listCfg.name {
				return nil
			}

			out := strings.Builder{}
			out.WriteString(styles.CatalogNameStyle.Render(catalog.Name) + " ")
			out.WriteString(styles.SchemaNameStyle.Render(meta.Schema) + " ")
			out.WriteString(styles.PackageNameStyle.Render(meta.Package) + " ")
			out.WriteString(styles.NameStyle.Render(meta.Name))
			out.WriteString("\n")
			fmt.Print(out.String())

			return nil
		})
		if err != nil {
			return fmt.Errorf("reading FBC for catalog %q: %w", catalog.Name, err)
		}
		rc.Close()
	}

	return nil
}
