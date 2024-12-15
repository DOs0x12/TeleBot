package storage

import (
	"fmt"
	"time"

	"github.com/DOs0x12/TeleBot/server/internal/entities/bot"
	"github.com/jackc/pgx"
)

type PgCommStorage struct {
	connection *pgx.ConnPool
}

const tableComm = `
CREATE TABLE IF NOT EXISTS commands (
	name varchar(max) PRIMARY KEY,
	description varchar(max) NOT NULL
)`

func NewPgCommStorage(address, database, user, pass string) (PgCommStorage, error) {
	connConfig := pgx.ConnConfig{Host: address, Database: database, User: user, Password: pass}
	conn, err := pgx.NewConnPool(pgx.ConnPoolConfig{ConnConfig: connConfig, AcquireTimeout: 5 * time.Second})
	if err != nil {
		return PgCommStorage{}, fmt.Errorf("can not connect to the storage server: %w", err)
	}

	_, err = conn.Exec(tableComm)
	if err != nil {
		return PgCommStorage{}, fmt.Errorf("can not create a table in the storage: %w", err)
	}

	return PgCommStorage{connection: conn}, nil
}

const selComm = `SELECT EXISTS(SELECT 1 FROM commands WHERE name=$1)`
const saveNewComm = `INSERT INTO commands (name, description) VALUES ($1, $2)`
const updateComm = `UPDATE commands SET description=$2 WHERE name=$1`

func (st PgCommStorage) Save(comm bot.Command) error {
	row := st.connection.QueryRow(selComm, comm.Name)
	var isCommExists bool
	err := row.Scan(&isCommExists)
	if err != nil {
		return fmt.Errorf("can not check if the command row exists in the storage: %w", err)
	}

	if isCommExists {
		_, err = st.connection.Exec(updateComm, comm.Name, comm.Description)
		if err != nil {
			return fmt.Errorf("can not update command data in the storage: %w", err)
		}

		return nil
	}

	_, err = st.connection.Exec(saveNewComm, comm.Name, comm.Description)
	if err != nil {
		return fmt.Errorf("can not write new command data into the storage: %w", err)
	}

	return nil
}

const loadCom = `SELECT name, description FROM commands`

func (st PgCommStorage) Load() ([]bot.Command, error) {
	rows, err := st.connection.Query(loadCom)
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
