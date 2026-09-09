package compute

import (
	"context"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	computeclient "go.mws.cloud/go-sdk/service/compute/client"
	computemodel "go.mws.cloud/go-sdk/service/compute/model"
	computesdk "go.mws.cloud/go-sdk/service/compute/sdk"

	"cq-source-mws/client"
)

func Images() *schema.Table {
	return &schema.Table{
		Name:        "mws_compute_images",
		Description: "Disk images of every synced project",
		Resolver:    fetchImages,
		Multiplex:   client.ProjectMultiplex(),
		Transform:   client.TransformResource(&computemodel.ImageOptionalResponse{}),
		Columns:     client.ProjectHierarchyColumns(),
	}
}

func fetchImages(ctx context.Context, meta schema.ClientMeta, _ *schema.Resource, res chan<- any) error {
	c := meta.(*client.Client)

	images, err := computesdk.NewImage(ctx, c.SDK)
	if err != nil {
		return err
	}

	req := computeclient.ListImagesRequest{Project: c.ProjectName}
	return client.List(ctx, req, images.ListImages, res, c.Skip(serviceName))
}
