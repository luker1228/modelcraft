package projectgraphql

import (
	"sync"

	appEnduser "modelcraft/internal/app/enduser"
)

var (
	endUserMgmtServiceMu sync.RWMutex
	endUserMgmtService   *appEnduser.EndUserManagementAppService
)

// SetEndUserManagementAppService sets the global end-user management service.
func SetEndUserManagementAppService(service *appEnduser.EndUserManagementAppService) {
	endUserMgmtServiceMu.Lock()
	defer endUserMgmtServiceMu.Unlock()
	endUserMgmtService = service
}

func getEndUserManagementAppService() *appEnduser.EndUserManagementAppService {
	endUserMgmtServiceMu.RLock()
	defer endUserMgmtServiceMu.RUnlock()
	return endUserMgmtService
}
