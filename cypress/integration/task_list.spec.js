describe("Work flow", () => {
  beforeEach(() => {
      cy.setCookie("Other", "other");
      cy.setCookie("XSRF-TOKEN", "abcde");
      cy.visit("/supervision/workflow");
  });

  it("has column headers", () => {
    cy.contains("Task type");
    cy.contains("Client");
    cy.contains("Case owner");
    cy.contains("Assigned to");
    cy.contains("Due date");
    cy.contains("Actions");
  })
});