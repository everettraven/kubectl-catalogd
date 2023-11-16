package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/alecthomas/chroma/quick"
	"github.com/everettraven/kubectl-catalogd/internal/fetch"
	"github.com/everettraven/kubectl-catalogd/internal/stream"
	"github.com/operator-framework/operator-registry/alpha/declcfg"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/yaml"
)

var inspectCmd = cobra.Command{
	Use:   "inspect [schema] [name] [flags]",
	Short: "Inspects catalog objects",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		inspectCfg.schema = args[0]
		inspectCfg.name = args[1]

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

		return inspect(fetcher, streamer, inspectCfg)
	},
}

type inspector struct {
	schema      string
	pkg         string
	name        string
	catalogName string
	output      string
	style       string
}

var inspectCfg = inspector{
	schema:      "",
	pkg:         "",
	name:        "",
	catalogName: "",
	output:      "",
	style:       "",
}

func init() {
	inspectCmd.Flags().StringVar(&inspectCfg.pkg, "package", "", "specify the FBC object package that should be used to filter the resulting output")
	inspectCmd.Flags().StringVar(&inspectCfg.catalogName, "catalog", "", "specify the catalog that should be used. By default it will fetch from all catalogs")
	inspectCmd.Flags().StringVar(&inspectCfg.output, "output", "json", "specify the output format. Valid values are 'json' and 'yaml'")
	inspectCmd.Flags().StringVar(&inspectCfg.style, "style", "", "specify the style to use for syntax highlighting. If this value is empty syntax highlighting is disabled.")
}

func inspect(fetcher fetch.CatalogFetcher, streamer stream.CatalogContentStreamer, inspectCfg inspector) error {
	ctx := context.Background()
	catalogs, err := fetcher.FetchCatalogs(ctx, fetch.WithNameFilter(inspectCfg.catalogName), fetch.WithUnpackedFilter())
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

			if inspectCfg.schema != "" && meta.Schema != inspectCfg.schema {
				return nil
			}

			if inspectCfg.pkg != "" && meta.Package != inspectCfg.pkg {
				return nil
			}

			if inspectCfg.name != "" && meta.Name != inspectCfg.name {
				return nil
			}

			outBytes, err := json.MarshalIndent(meta.Blob, "", "  ")
			if err != nil {
				return err
			}
			if inspectCfg.output == "yaml" {
				outBytes, err = yaml.JSONToYAML(outBytes)
				if err != nil {
					return err
				}
			}

			if inspectCfg.style != "" {
				return quick.Highlight(os.Stdout, string(outBytes), inspectCfg.output, "terminal16m", inspectCfg.style)
			}

			fmt.Print(string(outBytes))
			return nil
		})
		if err != nil {
			return fmt.Errorf("reading FBC for catalog %q: %w", catalog.Name, err)
		}
		rc.Close()
	}

	return nil
}
