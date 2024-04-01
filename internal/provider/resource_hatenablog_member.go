package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hatena/terraform-provider-hatenablog-members/internal/client"
)

type BlogMemberResource struct {
	client *client.Client
}

type memberResourceModel struct {
	Username types.String `tfsdk:"username"`
	Role     types.String `tfsdk:"role"`
}

// ensure that BlogMemberResource satisfies interfaces
var (
	_ resource.Resource = &BlogMemberResource{}
)

func NewBlogMemberResource() resource.Resource {
	return &BlogMemberResource{}
}

func (r *BlogMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_member"
}

func (r *BlogMemberResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Description: "The Hatena ID of the blog member.",
				Required:    true,
			},
			"role": schema.StringAttribute{
				Required:    true,
				Description: "Role of the blog member. Role must be one of 'admin'（管理者）, 'editor'（編集者）, or 'contributor'（寄稿者）.",
				Validators: []validator.String{
					stringvalidator.OneOf("admin", "editor", "contributor"),
				},
			},
		},
	}
}

func (r *BlogMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*blogMemberProviderData).Client
}

func (r *BlogMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan memberResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.AddMember(plan.Username.ValueString(), plan.Role.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to add member %s: %s", plan.Username.ValueString(), err))
		return
	}

	plan.Username = types.StringValue(res.Username)
	plan.Role = types.StringValue(res.Role)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *BlogMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state memberResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	members, err := r.client.ListMembers()
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to list members: %s", err.Error()))
		return
	}

	for _, member := range members {
		if member.Username == state.Username.ValueString() {

			diags = resp.State.Set(ctx, &memberResourceModel{
				Username: types.StringValue(member.Username),
				Role:     types.StringValue(member.Role),
			})
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	// member not found
	resp.State.RemoveResource(ctx)
}

func (r *BlogMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan memberResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.AddMember(plan.Username.ValueString(), plan.Role.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to update member %s: %s", plan.Username.ValueString(), err.Error()))
		return
	}

	plan.Username = types.StringValue(res.Username)
	plan.Role = types.StringValue(res.Role)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *BlogMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state memberResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Deleting member %s", state.Username.ValueString()))

	err := r.client.DeleteMember(state.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to remove member %s: %s", state.Username.ValueString(), err.Error()))
		return
	}

	resp.State.RemoveResource(ctx)
}
