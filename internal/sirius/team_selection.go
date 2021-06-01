package sirius

import (
	"encoding/json"
	"net/http"
)

type TeamMembers struct {
	TeamMembersId   int    `json:"id"`
	TeamMembersName string `json:"name"`
}

type TeamType struct {
	Handle string `json:"handle"`
	Label  string `json:"label"`
}

type TeamCollection struct {
	Id               int           `json:"id"`
	Members          []TeamMembers `json:"members"`
	Name             string        `json:"name"`
	UserSelectedTeam int
	SelectedTeamId   int
	Type             string
	TypeLabel        string
	TeamType         TeamType `json:"teamType"`
}

type TeamStoredData struct {
	TeamId       int
	SelectedTeam int
}

func (c *Client) GetTeamSelection(ctx Context, loggedInTeamId int, selectedTeamName int, selectedTeamMembers TeamSelected) ([]TeamCollection, error) {
	var v []TeamCollection
	var k TeamStoredData

	req, err := c.newRequest(ctx, http.MethodGet, "/api/v1/teams", nil)
	if err != nil {
		return v, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return v, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return v, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return v, newStatusError(resp)
	}

	if err = json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return v, err
	}

	if selectedTeamName == 0 && k.TeamId == 0 {
		k.TeamId = loggedInTeamId
	} else {
		k.TeamId = selectedTeamName
	}

	if selectedTeamMembers.selectedTeamToAssignTask == 0 && k.SelectedTeam == 0 {
		k.SelectedTeam = loggedInTeamId
	} else {
		k.SelectedTeam = selectedTeamMembers.selectedTeamToAssignTask
	}

	for i := range v {
		v[i].UserSelectedTeam = k.TeamId
		v[i].SelectedTeamId = k.SelectedTeam

	}
	v = filterOutNonLayTeams(v)

	return v, err
}

func filterOutNonLayTeams(v []TeamCollection) []TeamCollection {
	var filteredTeams []TeamCollection
	for _, s := range v {
		if len(s.TeamType.Handle) != 0 {
			filteredTeams = append(filteredTeams, s)
		}
	}
	return filteredTeams
}
