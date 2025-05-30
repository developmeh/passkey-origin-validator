# Project Development Guidelines

## Build Management
- This project builds are managed by make and the makefile should build on the run target
- After each task is complete run the application to make sure it builds using the `make run --debug`
- When the task is complete run the tests related to files that were updated

## Debugging and Testing
- You should introduce debug logging that is feature flagged and run with debug and it should evaluate the output against the task expectations.
- For each module generate tests before generating code for the solution. When tests are complete we can move on to building.

## Code Organization
- Modules should have single focused and expose interfaces.
- The program entry point should be kept in the cmd directory and implementation in internal while public interfaces should be exposed at the root.

## Documentation
- When building the project write a readme.md that explains the project and any dependencies
