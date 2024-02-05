package gomigrator

import (
	"context"
	"database/sql"
	"errors"
)

type Migrations []Migration

type Migration struct {
	id   int64
	name string

	isPrepared bool
	preparer   Preparer

	executors ExecutorContainer
}

type ExecutorContainer struct {
	noTx         bool
	executor     Executor
	executorNoTx ExecutorNoTx
}

type Executor interface {
	Up(ctx context.Context, tx *sql.Tx) error
	Down(ctx context.Context, tx *sql.Tx) error
}

type ExecutorNoTx interface {
	Up(ctx context.Context, db *sql.DB) error
	Down(ctx context.Context, db *sql.DB) error
}

type Preparer interface {
	Prepare() (*ExecutorContainer, error)
}

func NewExecutorContainer(executor Executor) *ExecutorContainer {
	return &ExecutorContainer{
		noTx:     false,
		executor: executor,
	}
}

func NewExecutorContainerNoTx(executorNoTx ExecutorNoTx) *ExecutorContainer {
	return &ExecutorContainer{
		noTx:         true,
		executorNoTx: executorNoTx,
	}
}

func (e ExecutorContainer) NoTx() bool {
	return e.noTx
}

func (e ExecutorContainer) Executor() Executor {
	return e.executor
}

func (e ExecutorContainer) ExecutorNoTx() ExecutorNoTx {
	return e.executorNoTx
}

func NewMigration(id int64, name string, executor Executor) Migration {
	return Migration{
		id:         id,
		name:       name,
		isPrepared: true,
		executors: ExecutorContainer{
			noTx:     false,
			executor: executor,
		},
	}
}

func NewMigrationNoTx(id int64, name string, executorNoTx ExecutorNoTx) Migration {
	return Migration{
		id:         id,
		name:       name,
		isPrepared: true,
		executors: ExecutorContainer{
			noTx:         true,
			executorNoTx: executorNoTx,
		},
	}
}

func NewMigrationWithPreparer(id int64, name string, preparer Preparer) Migration {
	return Migration{
		id:         id,
		name:       name,
		isPrepared: false,
		preparer:   preparer,
	}
}

var (
	ErrMigrationNotPrepared = errors.New("migration is not prepared")
	ErrNilMigrationExecutor = errors.New("migration executor is nil")
	ErrNilMigrationPreparer = errors.New("migration preparer is nil")
	ErrNilExecutorContainer = errors.New("migration preparer returned nil ExecutorContainer")
)

func (m *Migration) Up(ctx context.Context, tx *sql.Tx) error {
	if !m.isPrepared {
		return ErrMigrationNotPrepared
	}

	if m.executors.Executor() == nil {
		return ErrNilMigrationExecutor
	}

	return m.executors.Executor().Up(ctx, tx)
}

func (m *Migration) Down(ctx context.Context, tx *sql.Tx) error {
	if !m.isPrepared {
		return ErrMigrationNotPrepared
	}

	if m.executors.Executor() == nil {
		return ErrNilMigrationExecutor
	}

	return m.executors.Executor().Down(ctx, tx)
}

func (m *Migration) UpNoTx(ctx context.Context, db *sql.DB) error {
	if !m.isPrepared {
		return ErrMigrationNotPrepared
	}

	if m.executors.ExecutorNoTx() == nil {
		return ErrNilMigrationExecutor
	}

	return m.executors.ExecutorNoTx().Up(ctx, db)
}

func (m *Migration) DownNoTx(ctx context.Context, db *sql.DB) error {
	if !m.isPrepared {
		return ErrMigrationNotPrepared
	}

	if m.executors.ExecutorNoTx() == nil {
		return ErrNilMigrationExecutor
	}

	return m.executors.ExecutorNoTx().Down(ctx, db)
}

func (m *Migration) ChooseExecutor() (noTx bool, err error) {
	if err = m.ensureIsPrepared(); err != nil {
		return false, err
	}

	return m.executors.NoTx(), nil
}

func (m *Migration) ID() int64 {
	return m.id
}

func (m *Migration) Name() string {
	return m.name
}

func (m *Migration) ensureIsPrepared() error {
	if m.isPrepared {
		return nil
	}

	if m.preparer == nil {
		return ErrNilMigrationPreparer
	}

	return m.prepare()
}

func (m *Migration) prepare() error {
	executorController, err := m.preparer.Prepare()
	if err != nil {
		return err
	}

	if executorController == nil {
		return ErrNilExecutorContainer
	}

	m.executors = *executorController
	m.isPrepared = true

	return nil
}
