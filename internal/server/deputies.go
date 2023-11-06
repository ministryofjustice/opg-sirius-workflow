package server

import (
	"fmt"
	"github.com/ministryofjustice/opg-go-common/paginate"
	"github.com/ministryofjustice/opg-sirius-workflow/internal/model"
	"github.com/ministryofjustice/opg-sirius-workflow/internal/sirius"
	"github.com/ministryofjustice/opg-sirius-workflow/internal/urlbuilder"
	"net/http"
)

type DeputiesClient interface {
	GetDeputyList(sirius.Context, sirius.DeputyListParams) (sirius.DeputyList, error)
}

type DeputiesPage struct {
	DeputyList sirius.DeputyList
	ListPage
	FilterByECM
}

func (dp DeputiesPage) GetAppliedFilters() []string {
	var appliedFilters []string
	for _, u := range dp.ECMs {
		if u.IsSelected(dp.SelectedECMs) {
			appliedFilters = append(appliedFilters, u.Name)
		}
	}
	for _, s := range dp.SelectedECMs {
		if s == dp.NotAssignedTeamID {
			appliedFilters = append(appliedFilters, "Not Assigned")
		}
	}
	return appliedFilters
}

func (dp DeputiesPage) CreateUrlBuilder() urlbuilder.UrlBuilder {
	return urlbuilder.UrlBuilder{
		Path:            "deputies",
		SelectedTeam:    dp.App.SelectedTeam.Selector,
		SelectedPerPage: dp.PerPage,
		SelectedSort:    dp.Sort,
		SelectedFilters: []urlbuilder.Filter{
			urlbuilder.CreateFilter("ecm", dp.SelectedECMs, true),
		},
	}
}

func (dp DeputiesPage) getECMs(teams []model.Team, selectedTeam model.Team) []model.Assignee {
	var members []model.Assignee
	var deputyType string

	if selectedTeam.IsPro() {
		deputyType = "PRO"
	} else {
		deputyType = "PA"
	}
	
	for _, t := range teams {
		if t.Type == deputyType {
			for _, m := range t.Members {
				members = append(members, model.Assignee{
					Id:   m.Id,
					Name: m.Name,
				})
			}
		}
	}

	return members
}

func deputies(client DeputiesClient, tmpl Template) Handler {
	return func(app WorkflowVars, w http.ResponseWriter, r *http.Request) error {
		ctx := getContext(r)

		if r.Method != http.MethodGet {
			return StatusError(http.StatusMethodNotAllowed)
		}

		if !app.SelectedTeam.IsPro() && !app.SelectedTeam.IsPA() {
			page := ClientTasksPage{ListPage: ListPage{PerPage: 25}}
			return RedirectError(page.CreateUrlBuilder().GetTeamUrl(app.SelectedTeam))
		}

		params := r.URL.Query()
		page := paginate.GetRequestedPage(params.Get("page"))
		perPageOptions := []int{25, 50, 100}
		deputiesPerPage := paginate.GetRequestedElementsPerPage(params.Get("per-page"), perPageOptions)

		sort := urlbuilder.CreateSortFromURL(params, []string{"deputy", "noncompliance"})

		var selectedECMs []string
		if params.Has("ecm") {
			selectedECMs = params["ecm"]
		}

		deputyList, err := client.GetDeputyList(ctx, sirius.DeputyListParams{
			Team:         app.SelectedTeam,
			Page:         page,
			PerPage:      deputiesPerPage,
			Sort:         fmt.Sprintf("%s:%s", sort.OrderBy, sort.GetDirection()),
			SelectedECMs: selectedECMs,
		})
		if err != nil {
			return err
		}

		vars := DeputiesPage{
			DeputyList: deputyList,
		}

		vars.ECMs = vars.getECMs(app.Teams, app.SelectedTeam)
		vars.SelectedECMs = selectedECMs
		if app.SelectedTeam.IsPro() {
			vars.NotAssignedTeamID = app.EnvironmentVars.DefaultProTeamID
		} else {
			vars.NotAssignedTeamID = app.EnvironmentVars.DefaultPaTeamID
		}

		vars.PerPage = deputiesPerPage
		vars.Sort = sort
		vars.App = app
		vars.UrlBuilder = vars.CreateUrlBuilder()

		if page > deputyList.Pages.PageTotal && deputyList.Pages.PageTotal > 0 {
			return RedirectError(vars.UrlBuilder.GetPaginationUrl(deputyList.Pages.PageTotal, deputiesPerPage))
		}

		vars.Pagination = paginate.Pagination{
			CurrentPage:     deputyList.Pages.PageCurrent,
			TotalPages:      deputyList.Pages.PageTotal,
			TotalElements:   deputyList.TotalDeputies,
			ElementsPerPage: vars.PerPage,
			ElementName:     "deputies",
			PerPageOptions:  perPageOptions,
			UrlBuilder:      vars.UrlBuilder,
		}

		vars.AppliedFilters = vars.GetAppliedFilters()

		return tmpl.Execute(w, vars)
	}
}
