# AntiGravity Prompt: Phase 6 — SDKs & Adapters

## Context
Make Acumius integration trivial for developers.

## Your Task

### 1. Python SDK (`sdk/python/`)
Package: `acumius`. Modules: memory, trust, policy. Publish to PyPI.

### 2. TypeScript SDK (`sdk/typescript/`)
Package: `acumius`. Publish to npm.

### 3. LangGraph Adapter (`adapters/langgraph/`)
Implement LangGraph BaseStore interface.

### 4. CrewAI Adapter (`adapters/crewai/`)
CrewAI-compatible memory backend.

### 5. AutoGen Adapter (`adapters/autogen/`)
AutoGen-compatible memory backend.

### 6. OpenAI Adapter (`adapters/openai/`)
OpenAI Agents SDK-compatible.

### Acceptance Criteria
- [ ] pip install acumius works
- [ ] npm install acumius works
- [ ] LangGraph adapter passes tests
- [ ] Each adapter has working example
