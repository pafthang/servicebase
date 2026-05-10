package apis

import (
	"strings"

	recordmodels "github.com/pafthang/servicebase/services/record/models"
	"github.com/pafthang/servicebase/tools/search"
)

var ruleQueryParams = []string{search.FilterQueryParam, search.SortQueryParam}
var adminOnlyRuleFields = []string{"@collection.", "@request."}

func checkForAdminOnlyRuleFields(requestInfo *recordmodels.RequestInfo) error {
	if requestInfo == nil || requestInfo.AdminTeamAccess || len(requestInfo.Query) == 0 {
		return nil
	}

	for _, param := range ruleQueryParams {
		v, _ := requestInfo.Query[param].(string)
		if v == "" {
			continue
		}
		for _, field := range adminOnlyRuleFields {
			if strings.Contains(v, field) {
				return httpError(403, "Only admins can filter by "+field, nil)
			}
		}
	}

	return nil
}
