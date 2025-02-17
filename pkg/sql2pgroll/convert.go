// SPDX-License-Identifier: Apache-2.0

package sql2pgroll

import (
	"fmt"

	pgq "github.com/xataio/pg_query_go/v6"

	"github.com/xataio/pgroll/pkg/migrations"
)

var ErrStatementCount = fmt.Errorf("expected exactly one statement")

// Convert converts a SQL statement to a slice of pgroll operations.
func Convert(sql string) (migrations.Migration, error) {
	tree, err := pgq.Parse(sql)
	if err != nil {
		return migrations.Migration{}, fmt.Errorf("parse error: %w", err)
	}

	var mig migrations.Migration
	stmts := tree.GetStmts()
	for i, _ := range stmts {
		node := stmts[i].GetStmt().GetNode()

		var ops migrations.Operations
		switch node := (node).(type) {
		case *pgq.Node_CreateStmt:
			ops, err = convertCreateStmt(node.CreateStmt)
		case *pgq.Node_AlterTableStmt:
			ops, err = convertAlterTableStmt(node.AlterTableStmt)
			j := i + 1
			for j < len(stmts) {
				updateNode := stmts[j].GetStmt().GetNode()
				if u, ok := updateNode.(*pgq.Node_UpdateStmt); ok {
					colName, expr := getUpMigration(u.UpdateStmt)
					fmt.Println(colName, expr)
				}
			}
		case *pgq.Node_RenameStmt:
			ops, err = convertRenameStmt(node.RenameStmt)
		case *pgq.Node_DropStmt:
			ops, err = convertDropStatement(node.DropStmt)
		case *pgq.Node_IndexStmt:
			ops, err = convertCreateIndexStmt(node.IndexStmt)
		default:
			ops = makeRawSQLOperation(sql)
		}
		if err != nil {
			return migrations.Migration{
				Name:       mig.Name,
				Operations: makeRawSQLOperation(sql),
			}, err
		}
		mig.Operations = append(mig.Operations, ops...)
	}
	return mig, nil
}

func makeRawSQLOperation(sql string) migrations.Operations {
	return migrations.Operations{
		&migrations.OpRawSQL{Up: sql},
	}
}
