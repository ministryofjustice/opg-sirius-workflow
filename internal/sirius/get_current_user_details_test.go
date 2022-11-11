package sirius

import (
	"github.com/ministryofjustice/opg-go-common/logging"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

//func TestGetCurrentUserDetails(t *testing.T) {
//	pact := &dsl.Pact{
//		Consumer:          "sirius-workflow",
//		Provider:          "sirius",
//		Host:              "localhost",
//		PactFileWriteMode: "merge",
//		LogDir:            "../../logs",
//		PactDir:           "../../pacts",
//	}
//	defer pact.Teardown()
//
//	testCases := []struct {
//		name              string
//		setup             func()
//		cookies           []*http.Cookie
//		expectedMyDetails UserDetails
//		expectedError     error
//	}{
//		{
//			name: "Test Get Current Details",
//			setup: func() {
//				pact.
//					AddInteraction().
//					Given("I am a Lay Team user").
//					UponReceiving("A request to workflow page").
//					WithRequest(dsl.Request{
//						Method: http.MethodGet,
//						Path:   dsl.String("/api/v1/users/current"),
//						Headers: dsl.MapMatcher{
//							"X-XSRF-TOKEN":        dsl.String("abcde"),
//							"Cookie":              dsl.String("XSRF-TOKEN=abcde; Other=other"),
//							"OPG-Bypass-Membrane": dsl.String("1"),
//						},
//					}).
//					WillRespondWith(dsl.Response{
//						Status:  http.StatusOK,
//						Headers: dsl.MapMatcher{"Content-Type": dsl.String("application/json")},
//						Body: dsl.Like(map[string]interface{}{
//							"id":          dsl.Like(65),
//							"name":        dsl.Like("case"),
//							"phoneNumber": dsl.Like("12345678"),
//							"teams": dsl.EachLike(map[string]interface{}{
//								"displayName": dsl.Like("Lay Team 1 - (Supervision)"),
//								"id":          dsl.Like(13),
//							}, 1),
//							"displayName": dsl.Like("case manager"),
//							"deleted":     dsl.Like(false),
//							"email":       dsl.Like("case.manager@opgtest.com"),
//							"firstname":   dsl.Like("case"),
//							"surname":     dsl.Like("manager"),
//							"roles":       dsl.EachLike("Case Manager", 1),
//							"locked":      dsl.Like(false),
//							"suspended":   dsl.Like(false),
//						}),
//					})
//			},
//			cookies: []*http.Cookie{
//				{Name: "XSRF-TOKEN", Value: "abcde"},
//				{Name: "Other", Value: "other"},
//			},
//			expectedMyDetails: UserDetails{
//				ID:          65,
//				Name:        "case",
//				PhoneNumber: "12345678",
//				Teams: []MyDetailsTeam{
//					{
//						TeamId:      13,
//						DisplayName: "Lay Team 1 - (Supervision)",
//					},
//				},
//				DisplayName: "case manager",
//				Deleted:     false,
//				Email:       "case.manager@opgtest.com",
//				Firstname:   "case",
//				Surname:     "manager",
//				Roles:       []string{"Case Manager"},
//				Locked:      false,
//				Suspended:   false,
//			},
//		},
//
//		{
//			name: "Unauthorized",
//			setup: func() {
//				pact.
//					AddInteraction().
//					Given("I am a Lay Team user").
//					UponReceiving("A request to get my details without cookies").
//					WithRequest(dsl.Request{
//						Method: http.MethodGet,
//						Path:   dsl.String("/api/v1/users/current"),
//						Headers: dsl.MapMatcher{
//							"OPG-Bypass-Membrane": dsl.String("1"),
//						},
//					}).
//					WillRespondWith(dsl.Response{
//						Status: http.StatusUnauthorized,
//					})
//			},
//			expectedError: ErrUnauthorized,
//		},
//	}
//
//	for _, tc := range testCases {
//		t.Run(tc.name, func(t *testing.T) {
//			tc.setup()
//
//			assert.Nil(t, pact.Verify(func() error {
//				client, _ := NewClient(http.DefaultClient, fmt.Sprintf("http://localhost:%d", pact.Server.Port))
//
//				myDetails, err := client.GetCurrentUserDetails(getContext(tc.cookies))
//				assert.Equal(t, tc.expectedMyDetails, myDetails)
//				assert.Equal(t, tc.expectedError, err)
//				return nil
//			}))
//		})
//	}
//}

func TestMyDetailsStatusError(t *testing.T) {
	logger := logging.New(os.Stdout, "opg-sirius-workflow ")
	s := teapotServer()
	defer s.Close()

	client, _ := NewClient(http.DefaultClient, s.URL, logger)

	_, err := client.GetCurrentUserDetails(getContext(nil))
	assert.Equal(t, StatusError{
		Code:   http.StatusTeapot,
		URL:    s.URL + "/api/v1/users/current",
		Method: http.MethodGet,
	}, err)
}
