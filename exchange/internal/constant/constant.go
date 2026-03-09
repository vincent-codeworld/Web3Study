package constant

const (
	RoleTaker = "taker"
	RoleMaker = "maker"
)

const (
	CancelReasonDeducted   = "Self Trade Prevention Deducted"
	CancelReasonCancelNew  = "Self Trade Prevention Cancel New"
	CancelReasonCancelOld  = "Self Trade Prevention Cancel Old"
	CancelReasonCancelBoth = "Self Trade Prevention Cancel Both"
	CancelReasonPostOnly   = "Order Post Only"
	CancelReasonFillOrKill = "Order Fill Or Kill"
	CancelReasonIOC        = "Immediate or Cancel"
)
