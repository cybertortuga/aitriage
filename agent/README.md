# AITriage Agent — LangGraph Security Remediation

A Python LangGraph agent that connects to the AITriage Go security engine via MCP (Model Context Protocol) and orchestrates an autonomous, stateful **security remediation workflow**.

## Architecture

```
AITriage Go binary          Python LangGraph Agent
(serve --transport sse)  ←→  (graph.py)
         │                         │
    13 MCP tools               StateGraph
    ─────────────          ──────────────────
    aitriage_scan      →   scan → analyze → fix
    aitriage_secrets       → verify → (retry/done)
    aitriage_entropy       → report
    aitriage_architecture
    ...
```

## Quick Start

### 1. Start the AITriage MCP Server
```bash
# From the project root
./aitriage serve --transport sse --port 9090
```

### 2. Set up Python environment
```bash
cd agent/
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

### 3. Configure environment
```bash
# Required: API key for the LLM (pick one)
export GEMINI_API_KEY=your-key
# OR
export ANTHROPIC_API_KEY=your-key
# OR
export OPENAI_API_KEY=your-key

# Optional: LangSmith observability
export LANGCHAIN_TRACING_V2=true
export LANGCHAIN_API_KEY=your-langsmith-key
export LANGCHAIN_PROJECT=aitriage
```

### 4. Run the agent
```bash
# Scan and remediate the current directory
python main.py .

# Scan a specific project
python main.py /path/to/your/project

# Enable Human-in-the-Loop (asks before applying each fix)
AGENT_HITL=1 python main.py /path/to/project
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `AITRIAGE_MCP_URL` | `http://localhost:9090/sse` | AITriage MCP server URL |
| `AITRIAGE_MCP_TRANSPORT` | `sse` | `sse` or `stdio` |
| `AITRIAGE_MCP_BINARY` | `aitriage` | Binary path (for stdio mode) |
| `AGENT_MODEL` | `google-genai:gemini-2.5-flash` | LangChain model string |
| `AGENT_HITL` | `0` | `1` to enable Human-in-the-Loop |
| `LANGCHAIN_TRACING_V2` | `false` | Enable LangSmith tracing |
| `LANGCHAIN_API_KEY` | — | LangSmith API key |
| `LANGCHAIN_PROJECT` | `aitriage` | LangSmith project name |

## LangGraph Workflow

```
START
  │
  ▼
scan_node          Calls aitriage_scan via MCP
  │                Returns findings + SecurityScore
  ▼
analyze_node       LLM picks the highest-priority finding
  │
  ▼
fix_node  ◄───┐   LLM generates a code patch
  │           │   (max 3 attempts)
  ▼           │
verify_node   │   Re-scans to check if score improved
  │           │
  ├─ retry ───┘   Score didn't improve, try again
  │
  ├─ done         Score improved → report
  │
  └─ give_up      3 attempts failed → report with warning
  │
  ▼
report_node        Assembles markdown report
  │
 END
```

## Checkpointing

The agent uses SQLite checkpointing (`agent_state.db` by default). This means:
- Interrupted runs can be **resumed** with `--thread-id`
- Each project gets its own thread ID
- State survives crashes and restarts

```bash
# Resume an interrupted run
python main.py /project --thread-id abc123
```

## Docker Compose

Use the included `docker-compose.yaml` at the project root:

```bash
# Start both AITriage MCP server + Python agent
docker compose up aitriage-mcp aitriage-agent
```
