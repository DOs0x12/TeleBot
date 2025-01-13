package storage

import (
	"context"
	"fmt"

	"github.com/DOs0x12/TeleBot/server/v2/internal/entities/bot"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgCommStorage struct {
	connection *pgxpool.Pool
}

const tableComm = `
CREATE TABLE IF NOT EXISTS commands (
	name varchar(50) UNIQUE,
	description varchar(1000) NOT NULL,
	token varchar(100) NOT NULL
)`

type StorageConf struct {
	Address,
	Database,
	User,
	Pass string
}

func NewPgCommStorage(ctx context.Context, conf StorageConf) (PgCommStorage, error) {
	maxLifeTime := "8760h"
	sslMode := "disable"
	connStr := fmt.Sprintf(
		"host=%v dbname=%v user=%v password=%v pool_max_conn_lifetime=%v sslmode=%v",
		conf.Address, conf.Database, conf.User, conf.Pass, maxLifeTime, sslMode)
	connConf, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return PgCommStorage{}, fmt.Errorf("the config for the databes was not parsed: %v", err)
	}

	conn, err := pgxpool.NewWithConfig(ctx, connConf)
	if err != nil {
		return PgCommStorage{}, fmt.Errorf("failed to connect to the storage server: %w", err)
	}

	_, err = conn.Exec(ctx, tableComm)
	if err != nil {
		return PgCommStorage{}, fmt.Errorf("failed to create a table in the storage: %w", err)
	}

	return PgCommStorage{connection: conn}, nil
}

const selComm = `SELECT EXISTS(SELECT 1 FROM commands WHERE name=$1)`
const saveNewComm = `INSERT INTO commands (name, description, token) VALUES ($1, $2, $3)`
const updateComm = `UPDATE commands SET description=$2, token=$3 WHERE name=$1`

func (st PgCommStorage) Save(ctx context.Context, comm bot.Command) error {
	row := st.connection.QueryRow(ctx, selComm, comm.Name)
	var isCommExists bool
	err := row.Scan(&isCommExists)
	if err != nil {
		return fmt.Errorf("failed to check if the command row exists in the storage: %w", err)
	}

	if isCommExists {
		_, err = st.connection.Exec(ctx, updateComm, comm.Name, comm.Description, comm.Token)
		if err != nil {
			return fmt.Errorf("failed to update command data in the storage: %w", err)
		}

		return nil
	}

	_, err = st.connection.Exec(ctx, saveNewComm, comm.Name, comm.Description, comm.Token)
	if err != nil {
		return fmt.Errorf("failed to write new command data into the storage: %w", err)
	}

	return nil
}

const loadCom = `SELECT name, description, token FROM commands`

func (st PgCommStorage) Load(ctx context.Context) ([]bot.Command, error) {
	rows, err := st.connection.Query(ctx, loadCom)
	if err != nil {
		return nil, fmt.Errorf("failed to send a query to the storage: %w", err)
	}

	var commands []bot.Command
	var name string
	var descr string
	var token string

	for rows.Next() {
		err = rows.Scan(&name, &descr, &token)
		if err != nil {
			return nil, fmt.Errorf("failed to read the storage rows: %w", err)
		}

		commands = append(commands, bot.Command{Name: name, Description: descr, Token: token})
	}

	return commands, nil
}

func (st PgCommStorage) Close() {
	st.connection.Close()
}
