package provider

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/quota/armquota"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

type quotaLimitDataSource struct {
	quotaFactory *armquota.ClientFactory
}

type quotaLimitDataSourceModel struct {
	Scope             string          `tfsdk:"scope"`
	ResourceName      string          `tfsdk:"resource_name"`
	ID                *string         `tfsdk:"id"`
	Name              quotaLimitName  `tfsdk:"name"`
	IsQuotaApplicable bool            `tfsdk:"is_quota_applicable"`
	Limit             quotaLimitLimit `tfsdk:"limit"`
}

type quotaLimitName struct {
	LocalizedValue string `tfsdk:"localized_value"`
	Value          string `tfsdk:"value"`
}

type quotaLimitLimit struct {
	LimitObjectType string `tfsdk:"limit_object_type"`
	Value           int32  `tfsdk:"limit"`
	LimitType       string `tfsdk:"limit_type"`
}

var (
	_ datasource.DataSource              = &quotaLimitDataSource{}
	_ datasource.DataSourceWithConfigure = &quotaLimitDataSource{}
)

func NewQuotaLimitDataSource() datasource.DataSource {
	return &quotaLimitDataSource{}
}

func (d *quotaLimitDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	quotaFactory, ok := req.ProviderData.(*armquota.ClientFactory)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *armquota.ClientFactory, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.quotaFactory = quotaFactory
}

func (d *quotaLimitDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quota_limit"
}

func (d *quotaLimitDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"scope": schema.StringAttribute{
				Required: true,
			},
			"resource_name": schema.StringAttribute{
				Required: true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"localized_value": schema.StringAttribute{
							Computed: true,
						},
						"value": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			"is_quota_applicable": schema.BoolAttribute{
				Computed: true,
			},
			"limit": schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"limit_object_type": schema.StringAttribute{
							Computed: true,
						},
						"value": schema.Int32Attribute{
							Computed: true,
						},
						"limit_type": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *quotaLimitDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data quotaLimitDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	client := d.quotaFactory.NewClient()

	clientResp, err := client.Get(ctx, data.ResourceName, data.Scope, nil)

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Get Quota Limit",
			err.Error(),
		)
		return
	}

	quotaLimit := clientResp.CurrentQuotaLimitBase

	data.ID = quotaLimit.ID
	data.Name = quotaLimitName{
		LocalizedValue: *quotaLimit.Properties.Name.LocalizedValue,
		Value:          *quotaLimit.Properties.Name.Value,
	}
	data.IsQuotaApplicable = *quotaLimit.Properties.IsQuotaApplicable

	limit, ok := clientResp.Properties.Limit.(*armquota.LimitObject)

	if !ok {
		resp.Diagnostics.AddError(
			"Failed to Convert Limit",
			"Failed to Convert LimitJSONObject to LimitObject",
		)
		return
	}

	data.Limit = quotaLimitLimit{
		LimitObjectType: string(*limit.LimitObjectType),
		Value:           *limit.Value,
		LimitType:       data.Limit.LimitType,
	}
}
