package main

import (
	"database/sql"
	"github.com/Allen-Career-Institute/doordarshan-kendra/pkg/sfu"
)

type TenantLTClusterHandler struct {
	sfu.TenantDefaultClusterHandler
	Db *sql.DB
}

func (t *TenantLTClusterHandler) GetTenant() string {
	return "LT"
}
