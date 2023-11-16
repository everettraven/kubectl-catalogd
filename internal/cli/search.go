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

var searchCmd = cobra.Command{
	Use:   "search [input] [flags]",
	Short: "Searches catalog objects",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		searchCfg.query = args[0]

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
		return search(fetcher, streamer, searchCfg)
	},
}

type searcher struct {
	schema      string
	pkg         string
	catalogName string
	query       string
}

var searchCfg = searcher{
	schema:      "",
	pkg:         "",
	catalogName: "",
	query:       "",
}

func init() {
	searchCmd.Flags().StringVar(&searchCfg.schema, "schema", "", "specify the FBC object schema that should be used to filter the resulting output")
	searchCmd.Flags().StringVar(&searchCfg.pkg, "package", "", "specify the FBC object package that should be used to filter the resulting output")
	searchCmd.Flags().StringVar(&searchCfg.catalogName, "catalog", "", "specify the catalog that should be used. By default it will fetch from all catalogs")
}

func search(fetcher fetch.CatalogFetcher, streamer stream.CatalogContentStreamer, searchCfg searcher) error {
	ctx := context.Background()
	catalogs, err := fetcher.FetchCatalogs(ctx, fetch.WithNameFilter(searchCfg.catalogName), fetch.WithUnpackedFilter())
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

			if searchCfg.schema != "" && meta.Schema != searchCfg.schema {
				return nil
			}

			if searchCfg.pkg != "" && meta.Package != searchCfg.pkg {
				return nil
			}

			if !strings.Contains(meta.Name, searchCfg.query) {
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
