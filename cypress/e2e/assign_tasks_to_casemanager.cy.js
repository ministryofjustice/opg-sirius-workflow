describe("Reassign Tasks", () => {
    beforeEach(() => {
        cy.setCookie("Other", "other");
        cy.setCookie("XSRF-TOKEN", "abcde");
        cy.intercept('api/v1/teams/*', {
            body: {
                "members": [
                    {
                        "id": 76,
                        "displayName": "LayTeam1 User4",
                        "suspended": false,
                    },
                    {
                        "id": 75,
                        "displayName": "LayTeam1 User3",
                        "suspended": true,
                    },
                    {
                        "id": 74,
                        "displayName": "LayTeam1 User2",
                        "suspended": false,
                    },
                    {
                        "id": 73,
                        "displayName": "LayTeam1 User1",
                        "suspended": false,
                    }
                ]
            }})
        cy.visit("/supervision/workflow/1");
    });

    it("shows me a table of tasks", () => {
        cy.get(".govuk-table__body > :nth-child(1) > :nth-child(2) > .govuk-label").contains('Case work - Complaint review')
        cy.get(":nth-child(1) > :nth-child(3) > a").contains('Lizzo Surname')
        cy.get(":nth-child(1) > :nth-child(4) > .govuk-label").contains('Allocations - (Supervision)')
        cy.get(":nth-child(1) > :nth-child(5) > .govuk-label").contains('Allocations User3')
    });

    it("allows you to assign a task to a team and retains pagination and filters", () => {
        cy.visit('/supervision/workflow/1?testVar=testVal');
        cy.setCookie("success-route", "assignTasksToCasemanager");
        cy.get("#select-task-1").click()
        cy.get("#manage-task").should('be.visible').click()
        cy.get('.moj-manage-tasks__edit-panel > :nth-child(2)').should('be.visible').click()
        cy.get('#assignTeam').select('Pro Team 1 - (Supervision)')
        cy.intercept('PATCH', 'api/v1/users/*', {statusCode: 204})
        cy.get('#edit-save').click()
        cy.get("#success-banner").should('be.visible')
        cy.get("#success-banner").contains('1 tasks have been reassigned')
        cy.url().should('contain', '/supervision/workflow/1?testVar=testVal')
    });

    it("allows you to assign multiple tasks to an individual in a team", () => {
        cy.setCookie("success-route", "assignTasksToCasemanager");
        cy.get("#select-task-1").click()
        cy.get("#select-task-2").click()
        cy.get("#select-task-5").click()
        cy.get("#manage-task").should('be.visible').click()
        cy.get('.moj-manage-tasks__edit-panel > :nth-child(2)').should('be.visible').click()
        cy.get('#assignTeam').select('Pro Team 1 - (Supervision)');
        cy.intercept('PATCH', 'api/v1/users/*', {statusCode: 204})
        cy.get('#assignCM option:contains(LayTeam1 User3)').should('not.exist')
        cy.get('#assignCM option:contains(LayTeam1 User4)').should('exist')
        cy.get('#assignCM').select('LayTeam1 User4');
        cy.get('#edit-save').click()
        cy.get("#success-banner").should('be.visible')
        cy.get("#success-banner").contains('3 tasks have been reassigned')
    });

    it("can cancel out of reassigning a task", () => {
        cy.get("#select-task-1").check('1')
        cy.get("#manage-task").click()
        cy.get("#edit-cancel").click()
        cy.get(".moj-manage-tasks__edit-panel").should('not.be.visible')
    });
});