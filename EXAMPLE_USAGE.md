# Example: Using the Orchestration System

This document provides a concrete example of using `collaborate.sh` to orchestrate parallel AI agent work.

## Scenario

Let's say you want to implement a new feature: **"Add analytics dashboard with real-time metrics"**

## Step-by-Step Example

### 1. Review Context

First, ensure your `tasks.md` has relevant context:

```bash
cat tasks.md
```

This helps the orchestration system understand your project structure and requirements.

### 2. Run the Orchestration

```bash
./collaborate.sh "Add analytics dashboard with real-time metrics"
```

### 3. Expected Output

```
=========================================
  Collaborative AI Orchestration
=========================================

[INFO] Checking prerequisites...
[SUCCESS] Prerequisites check passed
[INFO] Splitting task: 'Add analytics dashboard with real-time metrics'
[INFO] Analyzing tasks.md for context...
[SUCCESS] Task split complete. Results in .collaborate_results/subtasks.json
[INFO] Starting parallel agent execution...
[INFO] [gemini] Starting sub-task 1: Analytics dashboard - Design and architecture
[INFO] [gemini] Sub-task 1 running with PID 12345
[INFO] [codex] Starting sub-task 2: Analytics dashboard - Core implementation
[INFO] [codex] Sub-task 2 running with PID 12346
[INFO] [amp] Starting sub-task 3: Analytics dashboard - Testing and validation
[INFO] [amp] Sub-task 3 running with PID 12347
[INFO] Waiting for all agents to complete...
[SUCCESS] All parallel agents completed
[INFO] Integrating results from all agents...
[INFO] Running opencode integration...
[SUCCESS] Integration complete!
[SUCCESS] Final output: .collaborate_results/integrated_output.md

[SUCCESS] Orchestration complete!

Results available in: .collaborate_results/
  - subtasks.json: Task breakdown
  - *_*.md: Individual agent outputs
  - integrated_output.md: Final integrated result
```

### 4. Review Results

#### Sub-tasks Created

```bash
cat .collaborate_results/subtasks.json
```

```json
{
  "subtasks": [
    {
      "id": 1,
      "description": "Design analytics dashboard architecture with component structure, data flow, and UI/UX specifications",
      "focus": "design"
    },
    {
      "id": 2,
      "description": "Implement analytics dashboard with real-time data fetching, state management, and visualization components",
      "focus": "implementation"
    },
    {
      "id": 3,
      "description": "Create comprehensive tests for analytics dashboard including unit tests, integration tests, and E2E scenarios",
      "focus": "testing"
    }
  ]
}
```

#### Agent Outputs

**Gemini (Design)**:
```bash
cat .collaborate_results/gemini_1.md
```

Sample output might include:
- Component architecture diagram
- Data flow specifications
- API endpoint designs
- UI/UX mockups (described)
- State management strategy

**Codex (Implementation)**:
```bash
cat .collaborate_results/codex_2.md
```

Sample output might include:
- React/Vue components
- API integration code
- State management setup
- Chart/visualization libraries integration
- Real-time WebSocket connections

**Claude (Testing)**:
```bash
cat .collaborate_results/amp_3.md
```

Sample output might include:
- Unit tests for components
- Integration tests for API calls
- E2E test scenarios
- Mock data generators
- Performance test cases

#### Integrated Result

```bash
cat .collaborate_results/integrated_output.md
```

This combines all three outputs into a cohesive implementation plan with:
- Unified architecture
- Complete code implementation
- Comprehensive test suite
- Deployment instructions
- Potential conflicts resolved

### 5. Apply the Results

Review and apply the integrated output to your project:

```bash
# Review the integrated output
less .collaborate_results/integrated_output.md

# Apply changes manually or use parts of the generated code
# (Always review before applying!)
```

## Smaller Example: Bug Fix

For a simpler task like fixing a bug:

```bash
./collaborate.sh "Fix OAuth token refresh not working after 1 hour"
```

**Resulting sub-tasks might be:**
1. **Design**: Root cause analysis and solution architecture
2. **Implementation**: Code changes to fix the refresh logic
3. **Testing**: Test cases to prevent regression

## Larger Example: New Module

For a complex task like adding an entire new module:

```bash
./collaborate.sh "Implement complete subscription management system with billing integration"
```

**Resulting sub-tasks might be:**
1. **Design**: System architecture, database schema, API contracts, payment gateway integration design
2. **Implementation**: Backend APIs, frontend UI, billing service integration, webhook handlers
3. **Testing**: Unit tests, integration tests, payment simulation, security tests

## Monitoring Progress

While the agents are working:

```bash
# In another terminal, watch progress
watch -n 2 'ls -lh .collaborate_results/'

# Or tail the logs
tail -f .collaborate_results/*.log
```

## Iterative Refinement

If the first run doesn't meet your needs:

```bash
# Review what was generated
cat .collaborate_results/integrated_output.md

# Re-run with more specific task description
./collaborate.sh "Add analytics dashboard with real-time metrics using Chart.js and WebSocket polling every 5 seconds"
```

## Real-World Workflow

Here's how you might use this in a real project:

```bash
# 1. Start with a feature branch
git checkout -b feature/analytics-dashboard

# 2. Run orchestration
./collaborate.sh "Add analytics dashboard with real-time metrics"

# 3. Review all outputs
ls .collaborate_results/

# 4. Start with design
cat .collaborate_results/gemini_1.md
# Implement based on design

# 5. Use implementation as reference
cat .collaborate_results/codex_2.md
# Copy relevant code, adapt to your project

# 6. Add tests
cat .collaborate_results/amp_3.md
# Implement suggested tests

# 7. Review integrated version
cat .collaborate_results/integrated_output.md
# Use this as your final guide

# 8. Commit incrementally
git add src/components/AnalyticsDashboard.jsx
git commit -m "Add analytics dashboard component structure"

# 9. Continue implementing...
```

## Tips for Better Results

### Be Specific

❌ Bad: "Make it better"
✅ Good: "Add error handling with retry logic and user-friendly error messages"

### Provide Context

❌ Bad: "Add authentication"
✅ Good: "Add OAuth 2.0 authentication using Google Sign-In, storing tokens securely in platform-specific storage"

### Use the tasks.md

The more context in `tasks.md`, the better the agents understand your project:

```bash
# Before running, update tasks.md with current status
vim tasks.md

# Then run orchestration
./collaborate.sh "Your task here"
```

## Comparing Agent Approaches

After orchestration completes, you can compare how different agents approached the same problem:

```bash
# See how Gemini designed it
cat .collaborate_results/gemini_1.md

# See how Codex implemented it
cat .collaborate_results/codex_2.md

# See how Claude tested it
cat .collaborate_results/amp_3.md

# See how they were integrated
cat .collaborate_results/integrated_output.md
```

This can give you multiple perspectives and let you cherry-pick the best ideas!

## Next Steps

After successful orchestration:

1. ✅ Review integrated output
2. ✅ Adapt code to your project structure
3. ✅ Run tests
4. ✅ Refactor as needed
5. ✅ Commit changes
6. ✅ Create PR or continue development

## When to Use Orchestration

**Good use cases:**
- New features with design, implementation, and testing components
- Complex refactoring that needs careful planning
- Adding new modules/subsystems
- Implementing specifications that need validation

**Not ideal for:**
- Simple one-line changes
- Quick bug fixes in a single file
- When you need immediate interactive feedback
- Tasks requiring deep context from many files

## Conclusion

The orchestration system is most powerful when you have:
- A clear task description
- Good context in `tasks.md`
- A task that benefits from multiple perspectives
- Time to review and integrate the results

Happy orchestrating! 🚀
