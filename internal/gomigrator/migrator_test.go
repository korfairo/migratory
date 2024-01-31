package gomigrator

import (
	"testing"

	"github.com/korfairo/migratory/internal/require"
)

func TestFindMissingMigrations(t *testing.T) {
	type args struct {
		migrations Migrations
		results    []MigrationResult
	}
	tests := map[string]struct {
		args        args
		wantMissing Migrations
		wantDirty   bool
	}{
		"all migrations missing": {
			args: args{
				migrations: allMigrations,
				results:    []MigrationResult{},
			},
			wantMissing: allMigrations,
			wantDirty:   false,
		},
		"no migrations missing": {
			args: args{
				migrations: allMigrations,
				results:    allResults,
			},
			wantMissing: nil,
			wantDirty:   false,
		},
		"half migrations missing": {
			args: args{
				migrations: allMigrations,
				results: []MigrationResult{
					{ID: 1},
					{ID: 2},
					{ID: 3},
					{ID: 4},
					{ID: 5},
				},
			},
			wantMissing: Migrations{
				Migration{id: 6},
				Migration{id: 7},
				Migration{id: 8},
				Migration{id: 9},
				Migration{id: 10},
			},
			wantDirty: false,
		},
		"last migration missing": {
			args: args{
				migrations: allMigrations,
				results: []MigrationResult{
					{ID: 1},
					{ID: 2},
					{ID: 3},
					{ID: 4},
					{ID: 5},
					{ID: 6},
					{ID: 7},
					{ID: 8},
					{ID: 9},
				},
			},
			wantMissing: Migrations{
				Migration{id: 10},
			},
			wantDirty: false,
		},
		"one dirty migration": {
			args: args{
				migrations: allMigrations,
				results: []MigrationResult{
					{ID: 2},
					{ID: 3},
					{ID: 4},
					{ID: 5},
					{ID: 6},
					{ID: 7},
					{ID: 8},
					{ID: 9},
					{ID: 10},
				},
			},
			wantMissing: Migrations{
				Migration{id: 1},
			},
			wantDirty: true,
		},
		"dirty migrations": {
			args: args{
				migrations: allMigrations,
				results: []MigrationResult{
					{ID: 2},
					{ID: 3},
					{ID: 5},
					{ID: 8},
				},
			},
			wantMissing: Migrations{
				Migration{id: 1},
				Migration{id: 4},
				Migration{id: 6},
				Migration{id: 7},
				Migration{id: 9},
				Migration{id: 10},
			},
			wantDirty: true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gotMissing, gotDirty := findMissingMigrations(test.args.migrations, test.args.results)
			require.Equal(t, gotMissing, test.wantMissing, "findMissingMigrations(...) missing migrations")
			require.Bool(t, gotDirty, test.wantDirty, "findMissingMigrations(...) dirty migrations found")
		})
	}
}

var allMigrations = Migrations{
	Migration{id: 1},
	Migration{id: 2},
	Migration{id: 3},
	Migration{id: 4},
	Migration{id: 5},
	Migration{id: 6},
	Migration{id: 7},
	Migration{id: 8},
	Migration{id: 9},
	Migration{id: 10},
}

var allResults = []MigrationResult{
	{ID: 1},
	{ID: 2},
	{ID: 3},
	{ID: 4},
	{ID: 5},
	{ID: 6},
	{ID: 7},
	{ID: 8},
	{ID: 9},
	{ID: 10},
}
