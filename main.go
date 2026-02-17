package main

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>
import (
	"github.com/Allen-Career-Institute/doordarshan-kendra/cmd/doordarshan-kendra"
	"github.com/Allen-Career-Institute/doordarshan-kendra/pkg/sfu"
)

func main() {
	envToConfigPathMap := map[string]string{
		"local":   "configs/local.env",
		"sandbox": "configs/dev.env",
		"dev":     "configs/dev.env",
		"stage":   "configs/stage.env",
		"prod":    "configs/prod.env",
	}
	tenantClusterHandlerMap := map[string]sfu.IClusterHandler{
		"MvN": &TenantMNClusterHandler{},
		"LT":  &TenantLTClusterHandler{},
	}
	cmd.Run(envToConfigPathMap, tenantClusterHandlerMap)
}
