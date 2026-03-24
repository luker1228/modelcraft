# AI Agents Documentation

This document provides guidelines for AI agents working with this project.

## Project Structure

This is the ModelCraft project with separate frontend and backend codebases:

- **Backend (Go)**: Code in `./modelcraft-go`. See [AGENTS.md](modelcraft-go/AGENTS.md)  
  @./modelcraft-go/AGENTS.md
- **Frontend**: Code in `./modelcraft-front`. See [AGENTS.md](modelcraft-front/AGENTS.md)  
  @./modelcraft-front/AGENTS.md

## No Absolute Paths

- Do not use absolute paths (e.g., `/root/modelcraft_project/...`). Always use relative paths (e.g., `./modelcraft-go/...`) when referencing files or directories.

## Use @ References for Documentation

> Refer to @docs/architecture.md for system flow.
or
- Refer to @docs/state-management.md before editing state.

Please refer to the respective documentation for detailed coding styles, patterns, and conventions for each component.
