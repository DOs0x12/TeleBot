package storage

import (
	"context"
	"fmt"

	"github.com/DOs0x12/TeleBot/server/internal/entities/bot"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgCommStorage struct {
	connection *pgxpool.Pool
}

const tableComm = `
CREATE TABLE IF NOT EXISTS commands (
	name varchar(50) UNIQUE,
	description varchar(1000) NOT NULL
)`

func NewPgCommStorage(ctx context.Context, address, database, user, pass string) (PgCommStorage, error) {
	maxLifeTime := "8760h"
	sslMode := "disable"
	connStr := fmt.Sprintf(
		"host=%v dbname=%v user=%v password=%v pool_max_conn_lifetime=%v sslmode=%v",
		address, database, user, pass, maxLifeTime, sslMode)
	connConf, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return PgCommStorage{}, fmt.Errorf("the config for the databes is not parsed: %v", err)
	}

	conn, err := pgxpool.NewWithConfig(ctx, connConf)
	if err != nil {
		return PgCommStorage{}, fmt.Errorf("can not connect to the storage server: %w", err)
	}

	_, err = conn.Exec(ctx, tableComm)
	if err != nil {
		return PgCommStorage{}, fmt.Errorf("can not create a table in the storage: %w", err)
	}

	return PgCommStorage{connection: conn}, nil
}

const selComm = `SELECT EXISTS(SELECT 1 FROM commands WHERE name=$1)`
const saveNewComm = `INSERT INTO commands (name, description) VALUES ($1, $2)`
const updateComm = `UPDATE commands SET description=$2 WHERE name=$1`

func (st PgCommStorage) Save(ctx context.Context, comm bot.Command) error {
	row := st.connection.QueryRow(ctx, selComm, comm.Name)
	var isCommExists bool
	err := row.Scan(&isCommExists)
	if err != nil {
		return fmt.Errorf("can not check if the command row exists in the storage: %w", err)
	}

	if isCommExists {
		_, err = st.connection.Exec(ctx, updateComm, comm.Name, comm.Description)
		if err != nil {
			return fmt.Errorf("can not update command data in the storage: %w", err)
		}

		return nil
	}

	_, err = st.connection.Exec(ctx, saveNewComm, comm.Name, comm.Description)
	if err != nil {
		return fmt.Errorf("can not write new command data into the storage: %w", err)
	}

	return nil
}

const loadCom = `SELECT name, description FROM commands`

func (st PgCommStorage) Load(ctx context.Context) ([]bot.Command, error) {
	rows, err := st.connection.Query(ctx, loadCom)
	if err != nil {
		return nil, fmt.Errorf("can not load commands from the storage: %w", err)
	}

	var commands []bot.Command
	var name string
	var descr string

	for rows.Next() {
		err = rows.Scan(&name, &descr)
		if err != nil {
			return nil, fmt.Errorf("can not read the storage rows: %w", err)
		}

		commands = append(commands, bot.Command{Name: name, Description: descr})
	}

	return commands, nil
}

func (st PgCommStorage) Close() {
	st.connection.Close()
}
