package aviatrix

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-aviatrix/goaviatrix"
)

func resourceAviatrixVGWConn() *schema.Resource {
	return &schema.Resource{
		Create: resourceAviatrixVGWConnCreate,
		Read:   resourceAviatrixVGWConnRead,
		Update: resourceAviatrixVGWConnUpdate,
		Delete: resourceAviatrixVGWConnDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"conn_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of for Transit GW to VGW connection connection which is going to be created.",
			},
			"gw_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Transit Gateway.",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "VPC-ID where the Transit Gateway is located.",
			},
			"bgp_vgw_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Id of AWS's VGW that is used for this connection.",
			},
			"bgp_local_as_num": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "BGP Local ASN (Autonomous System Number). Integer between 1-65535.",
			},
		},
	}
}

func resourceAviatrixVGWConnCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*goaviatrix.Client)
	vgwConn := &goaviatrix.VGWConn{
		ConnName:      d.Get("conn_name").(string),
		GwName:        d.Get("gw_name").(string),
		VPCId:         d.Get("vpc_id").(string),
		BgpVGWId:      d.Get("bgp_vgw_id").(string),
		BgpLocalAsNum: d.Get("bgp_local_as_num").(string),
	}

	log.Printf("[INFO] Creating Aviatrix VGW Connection: %#v", vgwConn)

	err := client.CreateVGWConn(vgwConn)
	if err != nil {
		return fmt.Errorf("failed to create Aviatrix VGWConn: %s", err)
	}
	d.SetId(vgwConn.ConnName + "~" + vgwConn.VPCId)
	return resourceAviatrixVGWConnRead(d, meta)
}

func resourceAviatrixVGWConnRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*goaviatrix.Client)

	connName := d.Get("conn_name").(string)
	vpcID := d.Get("vpc_id").(string)
	if connName == "" || vpcID == "" {
		id := d.Id()
		log.Printf("[DEBUG] Looks like an import, no connection name received. Import Id is %s", id)
		d.Set("conn_name", strings.Split(id, "~")[0])
		d.Set("vpc_id", strings.Split(id, "~")[1])
		d.SetId(id)
	}

	vgwConn := &goaviatrix.VGWConn{
		ConnName: d.Get("conn_name").(string),
		VPCId:    d.Get("vpc_id").(string),
	}
	vConn, err := client.GetVGWConnDetail(vgwConn)
	if err != nil {
		if err == goaviatrix.ErrNotFound {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("couldn't find Aviatrix VGW Connection: %s", err)
	}
	log.Printf("[INFO] Found Aviatrix VGW Connection: %#v", vConn)

	d.Set("conn_name", vConn.ConnName)
	d.Set("gw_name", vConn.GwName)
	d.Set("vpc_id", vConn.VPCId)
	d.Set("bgp_vgw_id", vConn.BgpVGWId)
	d.Set("bgp_local_as_num", vConn.BgpLocalAsNum)

	d.SetId(vConn.ConnName + "~" + vConn.VPCId)
	return nil
}

func resourceAviatrixVGWConnUpdate(d *schema.ResourceData, meta interface{}) error {
	return fmt.Errorf("aviatrix VGW Connection cannot be updated - delete and create a new one")
}

func resourceAviatrixVGWConnDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*goaviatrix.Client)
	vgwConn := &goaviatrix.VGWConn{
		ConnName: d.Get("conn_name").(string),
		VPCId:    d.Get("vpc_id").(string),
	}

	log.Printf("[INFO] Deleting Aviatrix vgw_conn: %#v", vgwConn)

	err := client.DeleteVGWConn(vgwConn)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return nil
		}
		return fmt.Errorf("failed to delete Aviatrix VGWConn: %s", err)
	}
	return nil
}
