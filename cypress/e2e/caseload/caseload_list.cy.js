describe("Caseload list", () => {
    before(() => {
        cy.setCookie("Other", "other");
        cy.setCookie("XSRF-TOKEN", "abcde");
        cy.visit("/caseload?team=21");
    });

    it("has column headers", () => {
        cy.get('[data-cy="Client"]').should("contain", "Client");
        cy.get('[data-cy="Report due date"]').should("contain", "Report due date");
        cy.get('[data-cy="Case owner"]').should("contain", "Case owner");
        cy.get('[data-cy="Supervision level"]').should("contain", "Supervision level");
        cy.get('[data-cy="Status"]').should("contain", "Status");
    })

    it("should have a table with the column Client", () => {
        cy.get('.govuk-table__body > .govuk-table__row > :nth-child(2)').should("contain", "Ro Bot")
    })

    it("should have a table with the column Report due date", () => {
        cy.get(".govuk-table__body > :nth-child(1) > :nth-child(3)").should("contain", "21/12/2023")
    })

    it("should have a table with the column Case owner", () => {
        cy.get(".govuk-table__body > :nth-child(1) > :nth-child(4)").should("contain", "Lay Team 1 - (Supervision)")
    })

    it("should have a table with the column Supervision level", () => {
        cy.get(".govuk-table__body > :nth-child(1) > :nth-child(5)").should("contain", "Minimal")
    })

    it("should have a table with the column Status", () => {
        cy.get(".govuk-table__body > :nth-child(1) > :nth-child(6)").should("contain", "Closed")
    })

    it("can cancel out of reassigning a client", () => {
        cy.get('#select-client-63').click();
        cy.get("#manage-client").should('be.visible').click();
        cy.get("#edit-cancel").click();
        cy.get(".moj-manage-list__edit-panel").should('not.be.visible');
    });

    it("allows you to reassign a client", () => {
        cy.intercept('api/v1/teams/21', {
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
        cy.visit("/caseload?team=21");
        cy.setCookie("success-route", "998");
        cy.get('#select-client-63').click();
        cy.get("#manage-client").should('be.visible').click();
        cy.get('.moj-manage-list__edit-panel > :nth-child(2)').should('be.visible').click()
        cy.get('#assignTeam').select('Lay Team 1 - (Supervision)');
        cy.intercept('PATCH', 'api/v1/users/*', {statusCode: 204})
        cy.get('#assignCM option:contains(LayTeam1 User3)').should('not.exist')
        cy.get('#assignCM option:contains(LayTeam1 User4)').should('exist')
        cy.get('#assignCM').select('LayTeam1 User4');
        cy.get('#edit-save').click()
        cy.get("#success-banner").should('be.visible')
        cy.get("#success-banner").contains('You have reassigned ')
    });
});
