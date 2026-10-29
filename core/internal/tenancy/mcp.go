package tenancy

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// G014 (docs/foundation/MCP.md): the transaction-scoped Group page of the MCP control surface's context.list_groups. Like
// the other transaction-scoped primitives it carries no actor and enforces no capability: the access facade decides
// membership and selected-Group validity on its facts first, in the same transaction, and row-level security admits only
// the context's rows. Only the access package may import tenancy.

// MaxSelectableGroupsPage bounds one page.
const MaxSelectableGroupsPage = 100

// SelectableGroupsPage reads at most limit+1 Groups the subject may select in Organization-only context, ordered by Group
// UUID after after (zero: from the start), so the caller can tell whether another page follows. manager selects the
// owner/admin form (every Group of org); otherwise the member form (the Groups with a binding of user).
func SelectableGroupsPage(ctx context.Context, tx pgx.Tx, org OrganizationID, user UserID, manager bool, after GroupID, limit int) ([]Group, error) {
	if zero(org) || zero(user) || limit < 1 || limit > MaxSelectableGroupsPage {
		return nil, newError(CodeInvalidID)
	}
	var (
		rows pgx.Rows
		err  error
	)
	if manager {
		rows, err = tx.Query(ctx, sqlSelectableGroupsManager, uuidArg(org), afterArg(after), limit+1)
	} else {
		rows, err = tx.Query(ctx, sqlSelectableGroupsMember, uuidArg(org), uuidArg(user), afterArg(after), limit+1)
	}
	out, err := collect(rows, err, scanGroup)
	if e := classify(err); e != nil {
		return nil, e
	}
	return out, nil
}

// afterArg is the page cursor: the all-zero UUID (below every canonical Group UUID) for the first page.
func afterArg(g GroupID) any { return uuidArg(g) }
