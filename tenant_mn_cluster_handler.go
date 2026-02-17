package main

import (
	"database/sql"
	"github.com/Allen-Career-Institute/doordarshan-kendra/pkg/sfu"
)

type TenantMNClusterHandler struct {
	sfu.TenantDefaultClusterHandler
	Db *sql.DB
}

func (t *TenantMNClusterHandler) GetTenant() string {
	return "MvN"
}
