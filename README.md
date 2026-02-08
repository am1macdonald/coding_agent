# Self-Building Coding Agent

A self-aware, self-improving coding agent that lives in its own directory.

## Current Capabilities

- Directory exploration (`ls.sh`)
- Code search (`rg.sh`)
- Self-configuration (`agent.py`)
- Status reporting

## Structure

```
./
├── agent.py          # Core agent system
├── agent_config.json # Agent configuration
├── ls.sh             # Directory listing tool
├── rg.sh             # Search tool (grep-based)
└── README.md         # This file
```

## Usage

```bash
# Run the agent
python3 agent.py

# Use tools directly
bash ls.sh
bash rg.sh <pattern> [path]
```

## Next Steps

- [ ] Add file editing capabilities
- [ ] Implement code generation
- [ ] Add testing framework
- [ ] Create self-update mechanism
- [ ] Add memory/learning system
