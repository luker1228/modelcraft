"""
GraphQL client utilities for Design-time tests.

This module provides utilities to create and configure GraphQL clients
for testing Design-time APIs (model design, cluster config, project management).

Runtime tests can also import and use these utilities.
"""

import json
import logging
from typing import Optional, Any, Dict
from gql import Client
from gql.transport.requests import RequestsHTTPTransport


# Create logger for GraphQL operations
logger = logging.getLogger("graphql_client")


class LoggingGraphQLClient:
    """
    GraphQL client wrapper with automatic request/response logging.

    This wrapper logs all GraphQL operations with their variables and responses
    for easier debugging and test traceability.

    Logs include:
    - Request: operation type, name, and variables
    - Response: data, errors, and extensions (if present)
    - HTTP details: status code, headers (optional)
    """

    def __init__(self, client: Client, verbose: bool = True):
        """
        Initialize logging wrapper.

        Args:
            client: Underlying GQL client instance
            verbose: Enable detailed logging (default: True)
        """
        self._client = client
        self._verbose = verbose

    def execute(self, document, variable_values: Optional[Dict[str, Any]] = None, *args, **kwargs):
        """
        Execute GraphQL query with automatic logging.

        Args:
            document: GraphQL document (gql query)
            variable_values: Query variables
            *args, **kwargs: Additional arguments passed to underlying client

        Returns:
            dict: Query result
        """
        if self._verbose:
            # Extract operation info from document
            operation_info = self._extract_operation_info(document)

            # Format request log
            separator = "=" * 80
            request_log = f"\n{separator}\n[GraphQL Request] {operation_info}\n{separator}"

            if variable_values:
                vars_str = json.dumps(variable_values, indent=2, ensure_ascii=False, default=str)
                request_log += f"\n[Variables]\n{vars_str}"

            # Log to both stdout and logging system
            print(request_log)
            logger.info(f"GraphQL Request: {operation_info}")
            if variable_values:
                logger.debug(f"Variables: {variable_values}")

        # Execute query
        try:
            execution_result = self._client.execute(
                document,
                variable_values=variable_values,
                get_execution_result=True,
                *args,
                **kwargs,
            )

            if self._verbose:
                response_log = f"\n[GraphQL Response] Success\n"
                response_log += json.dumps(execution_result.data, indent=2, ensure_ascii=False, default=str)
                if execution_result.extensions:
                    response_log += f"\n[Extensions]\n{json.dumps(execution_result.extensions, indent=2, ensure_ascii=False, default=str)}"
                response_log += f"\n{separator}\n"
                print(response_log)

            return execution_result.data

        except Exception as e:
            if self._verbose:
                error_msg = str(e)
                error_log = f"\n[GraphQL Response] Error: {type(e).__name__}\n"

                # Try to extract detailed error information
                if hasattr(e, 'errors'):
                    error_log += f"[Errors]\n{json.dumps(e.errors, indent=2, ensure_ascii=False, default=str)}\n"

                if hasattr(e, 'extensions'):
                    error_log += f"[Extensions]\n{json.dumps(e.extensions, indent=2, ensure_ascii=False, default=str)}\n"

                if hasattr(e, 'data'):
                    error_log += f"[Partial Data]\n{json.dumps(e.data, indent=2, ensure_ascii=False, default=str)}\n"

                # Always show the error message
                error_log += f"[Message]\n{error_msg}\n"
                error_log += f"{separator}\n"

                # Log to both stdout and logging system
                print(error_log)
                logger.error(f"GraphQL Error: {type(e).__name__}: {error_msg}")
                if hasattr(e, 'errors'):
                    logger.error(f"Errors: {e.errors}")
                if hasattr(e, 'extensions'):
                    logger.error(f"Extensions: {e.extensions}")
            raise

    def _extract_operation_info(self, document) -> str:
        """Extract operation info (type and name) from GraphQL document."""
        try:
            # Try to get the query string representation
            query_str = None
            if hasattr(document, 'loc') and hasattr(document.loc, 'source'):
                query_str = document.loc.source.body
            elif hasattr(document, '__str__'):
                query_str = str(document)

            if query_str:
                # Simple regex-like extraction of operation type and name
                import re
                # Match: mutation/query OperationName
                match = re.search(r'(mutation|query)\s+(\w+)', query_str, re.IGNORECASE)
                if match:
                    op_type = match.group(1).capitalize()
                    op_name = match.group(2)
                    return f"{op_type}: {op_name}"

                # Fallback: just get type
                match = re.search(r'(mutation|query)\s*\{', query_str, re.IGNORECASE)
                if match:
                    return f"{match.group(1).capitalize()}"

            # Try to access document structure
            doc = document
            if hasattr(doc, 'document_node'):
                doc = doc.document_node

            if hasattr(doc, 'definitions') and doc.definitions:
                definition = doc.definitions[0]

                op_type = "Operation"
                if hasattr(definition, 'operation'):
                    op_type = definition.operation.value.capitalize()

                if hasattr(definition, 'name') and definition.name:
                    op_name = definition.name.value
                    return f"{op_type}: {op_name}"

                # Get first field name
                if hasattr(definition, 'selection_set') and definition.selection_set:
                    selections = definition.selection_set.selections
                    if selections and hasattr(selections[0], 'name'):
                        field_name = selections[0].name.value
                        return f"{op_type}: {field_name}"

                return op_type

            return "GraphQL Operation"
        except Exception:
            return "GraphQL Operation"

    def __getattr__(self, name):
        """Delegate unknown attributes to underlying client."""
        return getattr(self._client, name)


def create_design_graphql_client(url: str, token: str = None, timeout: int = 30, verbose: bool = True) -> LoggingGraphQLClient:
    """
    Create a GraphQL client for Design-time API with automatic logging.

    Args:
        url: GraphQL endpoint URL (e.g., http://localhost:8080/org/modelcraft/design/graphql)
        token: Optional JWT access token for authenticated requests
        timeout: Request timeout in seconds (default: 30)
        verbose: Enable detailed request/response logging (default: True)

    Returns:
        LoggingGraphQLClient: Configured GQL client instance with logging

    Example:
        >>> from tests.design.common.graphql_client import create_design_graphql_client
        >>> client = create_design_graphql_client("http://localhost:8080/org/modelcraft/design/graphql")
        >>> result = client.execute(query)
        >>> # With authentication:
        >>> client = create_design_graphql_client(url, token="eyJ...")
        >>> # Disable verbose logging:
        >>> client = create_design_graphql_client(url, verbose=False)

    Note:
        The returned client will automatically log all GraphQL requests and responses.
        The logged response is the direct result from client.execute(), which includes
        the 'data' field and any 'errors' or 'extensions' if present.
    """
    headers = {}
    if token:
        headers["Authorization"] = f"Bearer {token}"

    # Use standard transport
    transport = RequestsHTTPTransport(
        url=url,
        headers=headers,
        verify=True,
        retries=3,
        timeout=timeout,
    )

    base_client = Client(
        transport=transport,
        fetch_schema_from_transport=False,
    )

    # Wrap with logging client
    return LoggingGraphQLClient(base_client, verbose=verbose)


def execute_graphql(client: LoggingGraphQLClient, query: str, variables: Optional[dict] = None) -> dict:
    """
    Execute a GraphQL query with error handling.

    Note: When using LoggingGraphQLClient, request/response logging is automatic.

    Args:
        client: GraphQL client instance (LoggingGraphQLClient)
        query: GraphQL query string
        variables: Optional query variables

    Returns:
        dict: Query result

    Raises:
        TransportQueryError: If GraphQL query fails
        Exception: For other errors

    Example:
        >>> result = execute_graphql(client, CREATE_PROJECT_QUERY, {"input": data})
    """
    from gql import gql

    try:
        result = client.execute(gql(query), variable_values=variables)
        return result
    except Exception as e:
        # Error details already logged by LoggingGraphQLClient
        raise
