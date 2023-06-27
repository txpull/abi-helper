package helpers

const (
	CONTRACT_PROCESS_STATUS_INIT = iota
	CONTRACT_PROCESS_STATUS_PENDING
	CONTRACT_PROCESS_STATUS_SUCCESS
	CONTRACT_PROCESS_STATUS_FAILED
)

func ContractProcessStatusToString(status int) string {
	switch status {
	case CONTRACT_PROCESS_STATUS_INIT:
		return "init"
	case CONTRACT_PROCESS_STATUS_PENDING:
		return "pending"
	case CONTRACT_PROCESS_STATUS_SUCCESS:
		return "success"
	case CONTRACT_PROCESS_STATUS_FAILED:
		return "failed"
	default:
		return "unknown"
	}
}
