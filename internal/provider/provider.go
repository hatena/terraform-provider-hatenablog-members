package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hatena/terraform-provider-hatenablog-members/internal/client"
)

type blogMemberProvider struct {
	version string
}

type blogMemberProviderModel struct {
	Username       types.String `tfsdk:"username"`
	Owner          types.String `tfsdk:"owner"`
	Apikey         types.String `tfsdk:"apikey"`
	BlogHost       types.String `tfsdk:"blog_host"`
	HatenablogHost types.String `tfsdk:"hatenablog_host"`
	Insecure       types.Bool   `tfsdk:"insecure"`
}

type blogMemberProviderData struct {
	Client *client.Client
}

// ensure that blogMemberProvider implements the provider.Provider interface
var (
	_ provider.Provider = &blogMemberProvider{}
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &blogMemberProvider{
			version: version,
		}
	}
}

func (p *blogMemberProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hatenablog-members"
	resp.Version = p.version
}

func (p *blogMemberProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A terraform provider which allows you to manage the members of a Hatena Blog.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Description: "The Hatena ID of the operator who owns the target blog or has administrative privileges for it.",
				Required:    true,
			},
			"owner": schema.StringAttribute{
				Description: "The Hatena ID of the owner of the target blog. If not specified, the value of 'username' will be used.",
				Optional:    true,
			},
			"apikey": schema.StringAttribute{
				Description: "The API key of the operator. Please visit https://blog.hatena.ne.jp/-/config to obtain your API key.",
				Required:    true,
				Sensitive:   true,
			},
			"blog_host": schema.StringAttribute{
				Description: "The domain name or host part of the target Hatena blog's URL. For regular blogs, this should be the domain name like 'staff.hatenablog.com'. If using the subdirectory option, this should be like '0123456789' for a blog URL like 'https://0123456789.hatenablog-oem.com'.",
				Required:    true,
			},
			"hatenablog_host": schema.StringAttribute{
				// for internal use
				Optional: true,
			},
			"insecure": schema.BoolAttribute{
				// for internal use
				Optional: true,
			},
		},
	}
}

func (p *blogMemberProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config blogMemberProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddError("username is required", "cannot use unknown value for username")
		return
	}
	username := config.Username.ValueString()
	if username == "" {
		resp.Diagnostics.AddError("username is required", "cannot use empty value for username")
		return
	}

	if config.Owner.IsUnknown() {
		resp.Diagnostics.AddError("owner is required", "cannot use unknown value for owner")
		return
	}
	owner := username
	if !config.Owner.IsNull() {
		owner = config.Owner.ValueString()
	}

	if config.Apikey.IsUnknown() {
		resp.Diagnostics.AddError("apikey is required", "cannot use unknown value for apikey")
		return
	}
	apikey := config.Apikey.ValueString()
	if apikey == "" {
		resp.Diagnostics.AddError("apikey is required", "cannot use empty value for apikey")
		return
	}

	if config.BlogHost.IsUnknown() {
		resp.Diagnostics.AddError("blog_host is required", "cannot use unknown value for blog_host")
		return
	}
	blogHost := config.BlogHost.ValueString()
	if blogHost == "" {
		resp.Diagnostics.AddError("blog_host is required", "cannot use empty value for blog_host")
		return
	}

	client := client.NewClient(p.version, username, apikey, owner, blogHost)

	if !config.HatenablogHost.IsNull() {
		client.SetHatenablogHost(config.HatenablogHost.ValueString())
	}
	if !config.HatenablogHost.IsNull() {
		client.SetInsecure(config.Insecure.ValueBool())
	}

	data := blogMemberProviderData{
		Client: client,
	}
	resp.DataSourceData = &data
	resp.ResourceData = &data
}

func (p *blogMemberProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewBlogMemberResource,
	}
}

func (p *blogMemberProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}
