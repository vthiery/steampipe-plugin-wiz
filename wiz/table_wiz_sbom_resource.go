package wiz

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v6/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
	"github.com/turbot/steampipe-plugin-wiz/api"
)

// qualPlaceholder is what Steampipe passes for a qual whose value is only known
// while the query runs, such as an uncorrelated subquery:
//
//	where type = (select 'CONTAINER_IMAGE')
//
// It is not a value the API can filter on, so it is not sent. Postgres applies
// the qual to the returned rows anyway.
const qualPlaceholder = "?"

//// TABLE DEFINITION

func tableWizSBOMResource(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "wiz_sbom_resource",
		Description: "Wiz SBOM Resource",
		List: &plugin.ListConfig{
			Hydrate: listWizSBOMResources,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "id", Require: plugin.Optional},
				{Name: "type", Require: plugin.Optional},
			},
		},
		Columns: []*plugin.Column{
			{Name: "name", Type: proto.ColumnType_STRING, Description: "The name of the resource, such as owner/repository/branch for a repository branch.", Transform: transform.FromField("Resource.Name")},
			{Name: "id", Type: proto.ColumnType_STRING, Description: "The Wiz identifier of the resource.", Transform: transform.FromField("Resource.Id")},
			{Name: "type", Type: proto.ColumnType_STRING, Description: "The type of the resource, such as REPOSITORY_BRANCH, CONTAINER_IMAGE, CONTAINER, VIRTUAL_MACHINE or SERVERLESS.", Transform: transform.FromField("Resource.Type")},
			{Name: "external_id", Type: proto.ColumnType_STRING, Description: "The identifier of the resource in its provider.", Transform: transform.FromField("Resource.ExternalId")},
			{Name: "native_type", Type: proto.ColumnType_STRING, Description: "The resource type as named by its provider.", Transform: transform.FromField("Resource.NativeType")},
			{Name: "cloud_platform", Type: proto.ColumnType_STRING, Description: "The platform hosting the resource, such as AWS, GCP, Kubernetes or GitHub.", Transform: transform.FromField("Resource.CloudPlatform")},
			{Name: "artifact_count", Type: proto.ColumnType_INT, Description: "The number of artifacts (packages, libraries, technologies) in the SBOM of the resource.", Transform: transform.FromField("Analytics.Artifacts.Count")},
			{Name: "version_count", Type: proto.ColumnType_INT, Description: "The number of distinct artifact versions in the SBOM of the resource.", Transform: transform.FromField("Analytics.Versions.Count")},
			{Name: "projects", Type: proto.ColumnType_JSON, Description: "The projects the resource belongs to.", Transform: transform.FromField("Resource.Projects")},
		},
	}
}

//// LIST FUNCTION

func listWizSBOMResources(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	conn, err := getClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("wiz_sbom_resource.listWizSBOMResources", "connection_error", err)
		return nil, err
	}

	options := &api.ListSBOMResourcesRequestConfiguration{}
	if id := d.EqualsQualString("id"); id == qualPlaceholder {
		plugin.Logger(ctx).Debug("wiz_sbom_resource.listWizSBOMResources", "id_qual", "placeholder, not sent to the API")
	} else {
		options.ResourceId = id
	}
	if resourceType := d.EqualsQualString("type"); resourceType == qualPlaceholder {
		plugin.Logger(ctx).Debug("wiz_sbom_resource.listWizSBOMResources", "type_qual", "placeholder, not sent to the API")
	} else {
		options.ResourceType = resourceType
	}

	pageLimit := 500
	limit := d.QueryContext.Limit
	if limit != nil && int(*limit) < pageLimit {
		pageLimit = int(*limit)
	}
	options.Limit = pageLimit

	for {
		query, err := api.ListSBOMResources(ctx, conn, options)
		if err != nil {
			plugin.Logger(ctx).Error("wiz_sbom_resource.listWizSBOMResources", "query_error", err)
			return nil, err
		}

		for _, resource := range query.SBOMArtifactsGroupedByValues.Nodes {
			if resource.Resource == nil {
				continue
			}
			d.StreamListItem(ctx, resource)

			// Context can be cancelled due to manual cancellation or the limit has been hit
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if !query.SBOMArtifactsGroupedByValues.PageInfo.HasNextPage {
			break
		}
		options.EndCursor = query.SBOMArtifactsGroupedByValues.PageInfo.EndCursor
	}

	return nil, nil
}
