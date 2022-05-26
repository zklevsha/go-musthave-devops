package db

import (
	"context"

	"github.com/jackc/pgx/v4"
)

type DBConnector struct {
	DSN string
}

func (d DBConnector) Avaliable() error {
	ctx := context.Background()
	con, err := pgx.Connect(ctx, d.DSN)
	if err == nil {
		defer con.Close(ctx)
		return con.Ping(ctx)
	}
	return err
}
