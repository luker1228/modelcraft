---
type: "query"
date: "2026-04-24T14:48:22.012864+00:00"
question: "Trace why executionContext bridges communities"
contributor: "graphify"
source_nodes: ["executionContext"]
---

# Q: Trace why executionContext bridges communities

## Answer

executionContext acts as the shared gqlgen runtime context, directly linked to resolver fieldContext methods and introspection paths across multiple schema domains, making it the highest-degree bridge node (1932 edges across 12 communities).

## Source Nodes

- executionContext