package main

import (
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/cmd/doordarshan-kendra"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/sfu"
)

func main() {
	envToConfigPathMap := map[string]string{
		"local": "configs/local.env",
	}
	tenantClusterHandlerMap := map[string]sfu.IClusterHandler{
		"default": &sfu.TenantDefaultClusterHandler{},
	}
	cmd.Run(envToConfigPathMap, tenantClusterHandlerMap)
}
