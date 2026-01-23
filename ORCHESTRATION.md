# AI Agent Orchestration System

This document describes how to use the `collaborate.sh` script to orchestrate multiple AI agents working in parallel on complex tasks.

## Overview

The orchestration system splits large tasks into smaller sub-tasks and runs multiple AI agents (Gemini, Codex/GPT-4, and Claude) in parallel to work on different aspects. The results are then integrated using OpenCode to produce a cohesive final output.

## Prerequisites

1. **OpenCode CLI**: Must be installed and accessible in your PATH
   ```bash
   # Check if opencode is installed
   which opencode
   ```

2. **jq** (optional but recommended): For better JSON parsing
   ```bash
   # macOS
   brew install jq
   
   # Linux
   apt-get install jq  # or yum install jq
   ```

3. **API Access**: Ensure you have access configured for the AI models you want to use

## Usage

### Basic Usage

```bash
./collaborate.sh "Your main task description"
```

### Example

```bash
./collaborate.sh "Build a web app with user authentication and dashboard"
```

## How It Works

### 1. Task Splitting

The script analyzes your main task and splits it into 3 focused sub-tasks:
- **Design Sub-task**: Architecture, design decisions, specifications
- **Implementation Sub-task**: Code, features, file creation
- **Testing Sub-task**: Tests, test cases, validation strategies

If `tasks.md` exists, it uses that context to create more intelligent splits.

### 2. Parallel Execution

Three agents work simultaneously:
- **Gemini**: Typically handles design/architecture
- **Codex (GPT-4)**: Typically handles implementation
- **Claude (Sonnet)**: Typically handles testing/validation

Each agent runs independently in the background, allowing true parallelism.

### 3. Integration

Once all agents complete, OpenCode integrates their outputs by:
- Reviewing all agent outputs
- Identifying conflicts or overlaps
- Merging the best ideas from each agent
- Creating a unified, production-ready output
- Highlighting any remaining gaps

## Output Structure

All results are stored in `.collaborate_results/`:

```
.collaborate_results/
├── subtasks.json              # Task breakdown
├── gemini_1.md                # Gemini's output for sub-task 1
├── codex_2.md                 # Codex's output for sub-task 2
├── amp_3.md                   # Claude's output for sub-task 3
├── integrated_output.md       # Final integrated result
└── *.log                      # Execution logs
```

## Configuration

### Environment Variables

```bash
# Set custom tasks file (default: tasks.md)
export TASKS_FILE="my-tasks.md"

# Set custom working directory (default: .)
export WORKDIR="/path/to/project"

# Run the script
./collaborate.sh "My task"
```

### Agent Models

Edit `collaborate.sh` to customize which models each agent uses:

```bash
# In the execute_subtask() function, modify the model names:
gemini)
    opencode --model "gemini-pro" ...
    ;;
codex)
    opencode --model "gpt-4" ...
    ;;
amp)
    opencode --model "claude-3-sonnet" ...
    ;;
```

## Advanced Usage

### Custom Agent Assignment

Modify the `run_parallel_agents()` function to assign specific sub-tasks to specific agents based on their strengths.

### Sequential Integration

If you need results to be integrated sequentially rather than all at once, modify the `integrate_results()` function.

### Adding More Agents

1. Add agent name to `AGENTS` array
2. Add case in `execute_subtask()` function
3. Adjust task splitting to create more sub-tasks

```bash
AGENTS=("gemini" "codex" "amp" "custom-agent")
```

## Example Workflow

```bash
# 1. Navigate to your project
cd /path/to/Google-Play-Developer-CLI

# 2. Ensure tasks.md is up to date with context
cat tasks.md

# 3. Run orchestration
./collaborate.sh "Implement OAuth 2.0 token refresh with retry logic"

# 4. Monitor progress
tail -f .collaborate_results/*.log

# 5. Review results
cat .collaborate_results/integrated_output.md

# 6. Apply the integrated output to your project
# (manually review and apply changes)
```

## Monitoring Execution

### Check Running Agents

```bash
# List all running opencode processes
ps aux | grep opencode

# Check specific agent PID files
cat .collaborate_results/*.pid
```

### View Real-time Logs

```bash
# Watch all logs
tail -f .collaborate_results/*.log

# Watch specific agent
tail -f .collaborate_results/gemini_1.log
```

## Troubleshooting

### Issue: "opencode CLI not found"

**Solution**: Install opencode or add it to your PATH

```bash
export PATH=$PATH:/path/to/opencode
```

### Issue: Agents taking too long

**Solution**: Check the logs for errors or API rate limiting

```bash
cat .collaborate_results/*.log
```

### Issue: Poor integration quality

**Solution**: 
1. Ensure `tasks.md` has sufficient context
2. Refine your main task description to be more specific
3. Manually review and adjust agent outputs before integration

### Issue: Conflicts in agent outputs

**Solution**: The integration step should resolve these, but you can:
1. Review individual agent outputs
2. Manually merge the best parts
3. Re-run with more specific sub-task descriptions

## Best Practices

1. **Clear Task Descriptions**: Be specific about what you want to achieve
2. **Context Files**: Keep `tasks.md` updated with relevant project context
3. **Review Outputs**: Always review integrated output before applying to production
4. **Iterative Refinement**: Run multiple times with refined prompts if needed
5. **Version Control**: Commit before running orchestration for easy rollback

## Performance Tips

- Agents run in parallel, so total time ≈ slowest agent
- Use more specific sub-tasks for faster, focused results
- Consider using faster models for non-critical sub-tasks
- Cache results in `.collaborate_results/` to avoid re-running

## Security Considerations

- Never commit `.collaborate_results/` with sensitive data
- Review all generated code before applying
- Agent outputs may contain suggestions that need validation
- Always test integrated code in a safe environment first

## Future Enhancements

Potential improvements to the orchestration system:

- [ ] Dynamic agent selection based on task type
- [ ] Recursive task splitting for very large tasks
- [ ] Progress indicators during parallel execution
- [ ] Automatic conflict resolution strategies
- [ ] Integration with CI/CD pipelines
- [ ] Result caching and incremental updates
- [ ] Multi-round refinement loops
- [ ] Agent performance metrics and selection

## Contributing

To improve the orchestration system:

1. Test with different task types
2. Document edge cases and solutions
3. Share successful agent configurations
4. Submit improvements to `collaborate.sh`

## License

Same as the parent project.
