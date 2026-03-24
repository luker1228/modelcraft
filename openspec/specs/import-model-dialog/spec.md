### Requirement: Import model button in sidebar
The system SHALL display a "导入模型" button in the model editor sidebar, adjacent to the existing "新建模型" button.

#### Scenario: Button is visible when a database is selected
- **WHEN** a user has selected a database in the model editor
- **THEN** the "导入模型" button SHALL be visible and enabled

### Requirement: Import model dialog shows available tables
The system SHALL display a dialog when "导入模型" is clicked that lists all database tables not yet imported as models.

#### Scenario: Dialog opens and fetches tables
- **WHEN** user clicks "导入模型"
- **THEN** a dialog opens and calls `listTables` for the currently selected database, showing a loading state until results arrive

#### Scenario: Tables are displayed as a searchable list
- **WHEN** the table list is loaded
- **THEN** the dialog displays all returned table names in a scrollable list with a search/filter input

#### Scenario: User can select a table
- **WHEN** user clicks a table name in the list
- **THEN** that table becomes selected (highlighted with background `#dadee5`) and the "导入" button becomes enabled

#### Scenario: No tables available
- **WHEN** `listTables` returns an empty list
- **THEN** the dialog displays an empty state message indicating all tables have already been imported

### Requirement: Import creates a model from the selected table
The system SHALL call `reverseEngineerModel` when the user confirms the import.

#### Scenario: Successful import
- **WHEN** user selects a table and clicks "导入"
- **THEN** the system calls `reverseEngineerModel` with the selected table name, closes the dialog, shows a success toast, and the new model appears in the sidebar model list

#### Scenario: Import shows loading state
- **WHEN** the import mutation is in flight
- **THEN** the "导入" button SHALL show a loading indicator and be disabled to prevent double submission

#### Scenario: Import fails with error
- **WHEN** `reverseEngineerModel` returns an error (e.g. model already exists)
- **THEN** the system displays the error message as a toast and the dialog remains open
