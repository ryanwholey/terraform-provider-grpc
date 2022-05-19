package provider

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/ryanwholey/terraform-provider-grpc/internal/grpcurl"
)

func dataSourceRequest() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRequestRead,

		Schema: map[string]*schema.Schema{
			"address": {
				Description: "The gRPC server address in 'host:port' format",
				Type:        schema.TypeString,
				Required:    true,
			},
			"method": {
				Description: "A fully qualified method name in 'service/method' or 'service.method' format",
				Type:        schema.TypeString,
				Required:    true,
			},
			"request_headers": {
				Description: "Headers added to the RPC request and the initial reflection request",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Default: map[string]interface{}{},
			},
			"body": {
				Description: "The RPC response",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"format": {
				Description:  "The requested format of the response",
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"json", "text"}, true),
				Default:      "json",
				Optional:     true,
			},
		},
	}
}

func dataSourceRequestRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	address := d.Get("address").(string)
	method := d.Get("method").(string)
	format := d.Get("format").(string)
	headers := d.Get("request_headers").(map[string]interface{})

	// TODO: Add user agent
	// TODO: Add TLC certs
	// TODO: Add insecure?
	client := grpcurl.New(address, nil)

	if err := client.Connect(ctx); err != nil {
		return diag.FromErr(err)
	}

	defer client.Close()

	headerList := []string{}
	for k, v := range headers {
		headerList = append(headerList, fmt.Sprintf("%s:%s", k, v.(string)))
	}

	b, err := client.InvokeRPC(ctx, method, headerList, "", grpcurl.InvokeRPCOptions{
		Format: format,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	hash := sha256.Sum256([]byte(fmt.Sprintf("%s/%s", address, method)))
	d.SetId(fmt.Sprintf("%x", hash[:]))

	if err := d.Set("body", string(b)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
