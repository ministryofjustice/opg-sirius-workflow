package server

import (
	"github.com/ministryofjustice/opg-sirius-workflow/internal/model"
	"github.com/ministryofjustice/opg-sirius-workflow/internal/paginate"
	"github.com/ministryofjustice/opg-sirius-workflow/internal/urlbuilder"
	"reflect"
)

type ListPage struct {
	App            WorkflowVars
	AppliedFilters []string
	Pagination     paginate.Pagination
	PerPage        int
	UrlBuilder     urlbuilder.UrlBuilder
}

type FilterByAssignee struct {
	ListPage
	AssigneeFilterName string
	SelectedAssignees  []string
	SelectedUnassigned string
}

type FilterByTaskType struct {
	ListPage
	TaskTypes         []model.TaskType
	SelectedTaskTypes []string
}

type FilterByDueDate struct {
	ListPage
	SelectedDueDateFrom string
	SelectedDueDateTo   string
}

type FilterByStatus struct {
	ListPage
	StatusOptions    []model.RefData
	SelectedStatuses []string
}

func (lp ListPage) HasFilterBy(page interface{}, filter string) bool {
	filters := map[string]interface{}{
		"assignee":  FilterByAssignee{},
		"due-date":  FilterByDueDate{},
		"status":    FilterByStatus{},
		"task-type": FilterByTaskType{},
	}

	extends := func(parent interface{}, child interface{}) bool {
		p := reflect.TypeOf(parent)
		c := reflect.TypeOf(child)
		for i := 0; i < p.NumField(); i++ {
			if f := p.Field(i); f.Type == c && f.Anonymous {
				return true
			}
		}
		return false
	}

	if f, ok := filters[filter]; ok {
		return extends(page, f)
	}
	return false
}
