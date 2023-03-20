package server

import (
	"fmt"
	"github.com/ministryofjustice/opg-go-common/logging"
	"net/http"
	"os"
	"strconv"

	"github.com/ministryofjustice/opg-sirius-workflow/internal/sirius"
)

type WorkflowInformation interface {
	GetCurrentUserDetails(sirius.Context) (sirius.UserDetails, error)
	GetTaskTypes(sirius.Context, []string) ([]sirius.ApiTaskTypes, error)
	GetTaskList(sirius.Context, int, int, int, []string, []sirius.ApiTaskTypes, []string) (sirius.TaskList, error)
	GetPageDetails(sirius.TaskList, int, int) sirius.PageDetails
	GetTeamsForSelection(sirius.Context, int, []string) ([]sirius.ReturnedTeamCollection, error)
	GetAssigneesForFilter(sirius.Context, int, []string) (sirius.AssigneesTeam, error)
	AssignTasksToCaseManager(sirius.Context, int, []string) error
	GetAppliedFilters(int, []sirius.ApiTaskTypes, []sirius.ReturnedTeamCollection, sirius.AssigneesTeam) []string
}

type workflowVars struct {
	Path           string
	XSRFToken      string
	MyDetails      sirius.UserDetails
	TaskList       sirius.TaskList
	PageDetails    sirius.PageDetails
	LoadTasks      []sirius.ApiTaskTypes
	TeamSelection  []sirius.ReturnedTeamCollection
	Assignees      sirius.AssigneesTeam
	AppliedFilters []string
	TeamIdFromForm int
	SuccessMessage string
	Error          string
	Errors         sirius.ValidationErrors
}

func checkForChangesToSelectedPagination(bothDisplayTaskLimits []string, currentTaskDisplayString string) int {
	currentTaskDisplay, _ := strconv.Atoi(currentTaskDisplayString)

	if len(bothDisplayTaskLimits) != 0 {
		topDisplayTaskLimit, _ := strconv.Atoi(bothDisplayTaskLimits[0])
		bottomDisplayTaskLimit, _ := strconv.Atoi(bothDisplayTaskLimits[1])
		if topDisplayTaskLimit != currentTaskDisplay {
			return topDisplayTaskLimit
		} else if bottomDisplayTaskLimit != currentTaskDisplay {
			return bottomDisplayTaskLimit
		} else {
			return currentTaskDisplay
		}
	}
	return 25
}

func getLoggedInTeam(myDetails sirius.UserDetails, defaultWorkflowTeam int) int {
	if len(myDetails.Teams) < 1 {
		return defaultWorkflowTeam
	} else {
		return myDetails.Teams[0].TeamId
	}
}

func getAssigneeIdForTask(logger *logging.Logger, teamId, assigneeId string) (int, error) {
	var assigneeIdForTask int
	var err error

	if assigneeId != "" {
		assigneeIdForTask, err = strconv.Atoi(assigneeId)
	} else if teamId != "" {
		assigneeIdForTask, err = strconv.Atoi(teamId)
	}
	if err != nil {
		logger.Print("getAssigneeIdForTask error: " + err.Error())
		return 0, err
	}
	return assigneeIdForTask, nil
}

func createTaskIdForUrl(taskIdArray []string) string {
	taskIdForUrl := ""

	for i := 0; i < len(taskIdArray); i++ {
		taskIdForUrl += taskIdArray[i]
		if i < (len(taskIdArray) - 1) {
			taskIdForUrl += "+"
		}
	}
	return taskIdForUrl
}

func getSelectedTeamId(r *http.Request, loggedInTeamId int) int {
	selectedTeamIdFromUrl, _ := strconv.Atoi(r.URL.Query().Get("change-team"))

	if selectedTeamIdFromUrl == 0 {
		selectedTeamIdFromForm, _ := strconv.Atoi(r.FormValue("change-team"))
		if selectedTeamIdFromForm == 0 {
			return loggedInTeamId
		}
		return selectedTeamIdFromForm
	}

	return selectedTeamIdFromUrl
}

func resetAssignees(urlSelectedTeamId int, selectedTeamId int, assigneeSelected []string) []string {
	if urlSelectedTeamId != selectedTeamId {
		return nil
	}
	return assigneeSelected
}

func changeSelectedTeamIdForForm(r *http.Request, selectedTeamId int) int {
	urlSelectedTeamId, _ := strconv.Atoi(r.URL.Query().Get("teamIdFromForm"))
	if urlSelectedTeamId == 0 {
		urlSelectedTeamId = selectedTeamId
	}
	return urlSelectedTeamId
}

func loggingInfoForWorkflow(client WorkflowInformation, tmpl Template, defaultWorkflowTeam int) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		logger := logging.New(os.Stdout, "opg-sirius-workflow ")
		ctx := getContext(r)
		search, _ := strconv.Atoi(r.FormValue("page"))
		if search < 1 {
			search = 1
		}

		err := r.ParseForm()
		if err != nil {
			logger.Print("ParseForm error: " + err.Error())
			return err
		}

		displayTaskLimit := checkForChangesToSelectedPagination(r.Form["tasksPerPage"], r.FormValue("currentTaskDisplay"))

		taskTypeSelected := r.Form["selected-task-type"]
		assigneeSelected := r.Form["selected-assignee"]

		myDetails, err := client.GetCurrentUserDetails(ctx)
		if err != nil {
			logger.Print("GetCurrentUserDetails error " + err.Error())
			return err
		}

		loggedInTeamId := getLoggedInTeam(myDetails, defaultWorkflowTeam)
		selectedTeamId := getSelectedTeamId(r, loggedInTeamId)
		urlSelectedTeamId := changeSelectedTeamIdForForm(r, selectedTeamId)
		assigneeSelected = resetAssignees(urlSelectedTeamId, selectedTeamId, assigneeSelected)

		loadTaskTypes, err := client.GetTaskTypes(ctx, taskTypeSelected)
		if err != nil {
			logger.Print("GetTaskTypes error " + err.Error())
			return err
		}

		taskList, err := client.GetTaskList(ctx, search, displayTaskLimit, selectedTeamId, taskTypeSelected, loadTaskTypes, assigneeSelected)

		if err != nil {
			logger.Print("GetTaskList error " + err.Error())
			return err
		}
		if search > taskList.Pages.PageTotal && taskList.Pages.PageTotal > 0 {
			search = taskList.Pages.PageTotal
			taskList, err = client.GetTaskList(ctx, search, displayTaskLimit, selectedTeamId, taskTypeSelected, loadTaskTypes, assigneeSelected)

			if err != nil {
				logger.Print("GetTaskList error " + err.Error())
				return err
			}
		}

		pageDetails := client.GetPageDetails(taskList, search, displayTaskLimit)

		teamSelection, err := client.GetTeamsForSelection(ctx, selectedTeamId, assigneeSelected)

		if err != nil {
			logger.Print("GetTeamsForSelection error " + err.Error())
			return err
		}

		assigneesForFilter, err := client.GetAssigneesForFilter(ctx, selectedTeamId, assigneeSelected)

		if err != nil {
			logger.Print("GetAssigneesForFilter error " + err.Error())
			return err
		}

		appliedFilters := client.GetAppliedFilters(selectedTeamId, loadTaskTypes, teamSelection, assigneesForFilter)

		vars := workflowVars{
			Path:           r.URL.Path,
			XSRFToken:      ctx.XSRFToken,
			MyDetails:      myDetails,
			TaskList:       taskList,
			PageDetails:    pageDetails,
			LoadTasks:      loadTaskTypes,
			TeamSelection:  teamSelection,
			Assignees:      assigneesForFilter,
			AppliedFilters: appliedFilters,
			TeamIdFromForm: selectedTeamId,
		}

		if err != nil {
			return StatusError(http.StatusNotFound)
		}

		switch r.Method {
		case http.MethodGet:
			return tmpl.ExecuteTemplate(w, "page", vars)
		case http.MethodPost:
			var newAssigneeIdForTask int
			selectedTeamToAssignTaskString := r.FormValue("assignTeam")
			if selectedTeamToAssignTaskString == "0" {
				vars.Errors = sirius.ValidationErrors{
					"selection": map[string]string{"": "Please select a team"},
				}

				return tmpl.ExecuteTemplate(w, "page", vars)
			}
			//this is where it picks up the new user to assign task to
			newAssigneeIdForTask, err = getAssigneeIdForTask(logger, selectedTeamToAssignTaskString, r.FormValue("assignCM"))
			if err != nil {
				logger.Print("getAssigneeIdForTask error: " + err.Error())
				return err
			}

			err := r.ParseForm()
			if err != nil {
				logger.Print("ParseForm error: " + err.Error())
				return err
			}

			taskIdArray := (r.Form["selected-tasks"])
			//taskIdForUrl := createTaskIdForUrl(taskIdArray)

			if err != nil {
				logger.Print("taskIdForUrl error: " + err.Error())
				return err
			}

			// Attempt to save
			err = client.AssignTasksToCaseManager(ctx, newAssigneeIdForTask, taskIdArray)
			if err != nil {
				logger.Print("AssignTasksToCaseManager: " + err.Error())
				return err
			}

			if vars.Errors == nil {
				vars.SuccessMessage = fmt.Sprintf("%d tasks have been reassigned", len(taskIdArray))
			}

			vars.TaskList, err = client.GetTaskList(ctx, search, displayTaskLimit, selectedTeamId, taskTypeSelected, loadTaskTypes, assigneeSelected)
			if err != nil {
				logger.Print("vars.TaskList error: " + err.Error())
				return err
			}

			vars.PageDetails = client.GetPageDetails(vars.TaskList, search, displayTaskLimit)

			return tmpl.ExecuteTemplate(w, "page", vars)
		default:
			return StatusError(http.StatusMethodNotAllowed)
		}
	}
}
