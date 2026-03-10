#!/bin/bash

# collaborate.sh - Orchestrate parallel AI agent collaboration
# Usage: ./collaborate.sh "Main task description"

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
TASKS_FILE="${TASKS_FILE:-tasks.md}"
WORKDIR="${WORKDIR:-.}"
RESULTS_DIR="${WORKDIR}/.collaborate_results"
AGENTS=("gemini" "codex" "amp")

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    if ! command -v opencode &> /dev/null; then
        log_error "opencode CLI not found. Please install it first."
        exit 1
    fi
    
    if [ ! -f "$TASKS_FILE" ]; then
        log_warning "Tasks file not found: $TASKS_FILE"
        log_info "Using generic task splitting instead"
    fi
    
    log_success "Prerequisites check passed"
}

# Extract pending tasks from tasks.md
extract_pending_tasks() {
    local tasks_file="$1"
    
    if [ ! -f "$tasks_file" ]; then
        echo "[]"
        return
    fi
    
    # Extract all pending tasks (lines with [ ])
    grep -n "^[[:space:]]*- \[ \]" "$tasks_file" | while IFS=: read -r line_num content; do
        # Extract task description and any sub-bullets
        task_desc=$(echo "$content" | sed 's/^[[:space:]]*- \[ \][[:space:]]*//' | sed 's/[[:space:]]*$//')
        echo "$line_num:$task_desc"
    done
}

# Analyze task context from tasks.md
analyze_task_context() {
    local main_task="$1"
    local tasks_file="$2"
    
    log_info "Analyzing $tasks_file to find context for: '$main_task'"
    
    # Find the most relevant phase/section
    # Look for the task in tasks.md or related keywords
    cat > "$RESULTS_DIR/analysis_prompt.txt" <<EOF
You are analyzing the tasks.md file for a Google Play Developer CLI project.

Task to work on: $main_task

Please analyze the tasks.md file and provide:
1. Which phase this task belongs to (or create new if needed)
2. What are the dependencies (what must be done first)
3. What files/components will likely be affected
4. Any specific requirements mentioned

Output as JSON:
{
  "phase": "Phase name",
  "dependencies": ["dependency 1", "dependency 2"],
  "affected_components": ["component1", "component2"],
  "requirements": ["req 1", "req 2"],
  "context_summary": "Brief summary of relevant context"
}
EOF

    opencode --files "$tasks_file" < "$RESULTS_DIR/analysis_prompt.txt" > "$RESULTS_DIR/task_context.json" 2>&1
    
    log_success "Context analysis complete"
}

# Split main task into sub-tasks based on tasks.md structure
split_task() {
    local main_task="$1"
    local tasks_file="$2"
    
    log_info "Splitting task: '$main_task'"
    
    # Create results directory
    mkdir -p "$RESULTS_DIR"
    
    # Check if this is a pending task from tasks.md
    local is_from_tasks=false
    if [ -f "$tasks_file" ]; then
        if grep -q "$main_task" "$tasks_file"; then
            is_from_tasks=true
            log_info "Task found in $tasks_file - using context-aware splitting"
        fi
    fi
    
    # If tasks.md exists, analyze it to extract relevant sub-tasks
    if [ -f "$tasks_file" ]; then
        log_info "Analyzing $tasks_file for context..."
        
        # First, analyze the context
        analyze_task_context "$main_task" "$tasks_file"
        
        # Use opencode to analyze tasks and split the work
        cat > "$RESULTS_DIR/split_prompt.txt" <<EOF
You are orchestrating work on the Google Play Developer CLI project.

Main Task: $main_task

Based on the tasks.md file (which contains the implementation plan), create 3 focused sub-tasks 
that can be worked on in parallel. Consider:

1. The task structure in tasks.md (phases, dependencies, requirements)
2. Whether this is a new task or completing an existing pending task
3. The vertical slice approach mentioned in tasks.md
4. Testing requirements (property tests, unit tests, integration tests)

Create sub-tasks that align with the project's patterns:
- One focused on design/architecture/contracts
- One focused on implementation/code
- One focused on testing/validation

Output exactly 3 sub-tasks in JSON format:
{
  "subtasks": [
    {
      "id": 1, 
      "description": "Detailed description of first sub-task",
      "focus": "design",
      "files_to_create_or_modify": ["file1.go", "file2.go"],
      "requirements_refs": ["1.1", "1.2"]
    },
    {
      "id": 2,
      "description": "Detailed description of second sub-task", 
      "focus": "implementation",
      "files_to_create_or_modify": ["file3.go", "file4.go"],
      "requirements_refs": ["1.3", "1.4"]
    },
    {
      "id": 3,
      "description": "Detailed description of third sub-task",
      "focus": "testing",
      "files_to_create_or_modify": ["file1_test.go", "file2_test.go"],
      "requirements_refs": ["testing"]
    }
  ],
  "phase": "Phase X",
  "dependencies": ["completed task 1", "completed task 2"]
}
EOF
        
        # Run opencode to split the task
        opencode --files "$tasks_file" < "$RESULTS_DIR/split_prompt.txt" > "$RESULTS_DIR/subtasks.json" 2>&1
        
    else
        # Generic task splitting
        log_info "Performing generic task split into design, implementation, and testing..."
        
        cat > "$RESULTS_DIR/subtasks.json" <<EOF
{
  "subtasks": [
    {"id": 1, "description": "$main_task - Design and architecture", "focus": "design", "files_to_create_or_modify": [], "requirements_refs": []},
    {"id": 2, "description": "$main_task - Core implementation", "focus": "implementation", "files_to_create_or_modify": [], "requirements_refs": []},
    {"id": 3, "description": "$main_task - Testing and validation", "focus": "testing", "files_to_create_or_modify": [], "requirements_refs": []}
  ],
  "phase": "Custom",
  "dependencies": []
}
EOF
    fi
    
    log_success "Task split complete. Results in $RESULTS_DIR/subtasks.json"
}

# Execute a sub-task with a specific agent
execute_subtask() {
    local agent="$1"
    local subtask_id="$2"
    local subtask_desc="$3"
    local focus="$4"
    local files_to_modify="${5:-}"
    local requirements="${6:-}"
    
    log_info "[$agent] Starting sub-task $subtask_id: $subtask_desc"
    
    local output_file="$RESULTS_DIR/${agent}_${subtask_id}.md"
    local log_file="$RESULTS_DIR/${agent}_${subtask_id}.log"
    
    # Create prompt for the agent based on Google Play CLI context
    cat > "$RESULTS_DIR/${agent}_${subtask_id}_prompt.txt" <<EOF
You are working on the Google Play Developer CLI project, a Go-based tool for managing apps on Google Play.

Focus Area: $focus
Sub-task: $subtask_desc
EOF

    # Add files context if available
    if [ -n "$files_to_modify" ] && [ "$files_to_modify" != "null" ]; then
        cat >> "$RESULTS_DIR/${agent}_${subtask_id}_prompt.txt" <<EOF

Files to create or modify: $files_to_modify
EOF
    fi

    # Add requirements context if available
    if [ -n "$requirements" ] && [ "$requirements" != "null" ]; then
        cat >> "$RESULTS_DIR/${agent}_${subtask_id}_prompt.txt" <<EOF

Related requirements: $requirements
EOF
    fi

    # Add role-specific guidance
    cat >> "$RESULTS_DIR/${agent}_${subtask_id}_prompt.txt" <<EOF

Your role based on focus area:

DESIGN (architecture, contracts, specifications):
- Define data structures, interfaces, and contracts
- Specify JSON output formats and error schemas
- Design command-line flags and arguments
- Document configuration and behavior
- Reference the specification in design.md and requirements.md

IMPLEMENTATION (code, features, logic):
- Write Go code following the project structure (cmd/, internal/, pkg/)
- Implement commands in the appropriate namespace (publish/*, reviews/*, etc.)
- Follow the Result envelope pattern for all commands
- Use the established error handling and exit codes
- Integrate with Google Play APIs (Android Publisher v3, Play Developer Reporting)
- Follow idempotency patterns where applicable

TESTING (validation, test cases, property tests):
- Write property tests for critical contracts (as outlined in tasks.md)
- Create unit tests with table-driven test patterns
- Use the mock transport harness for API testing
- Create or update golden fixtures for request/response validation
- Ensure tests cover edge cases and error conditions
- Aim for comprehensive coverage of the implemented functionality

Project context:
- This is a CLI tool wrapping Google Play Developer API
- All commands return JSON output with metadata envelope
- Uses Go 1.21+ with standard library patterns
- Follows vertical slice approach: complete minimal features first
- Strong emphasis on contract stability and testability

Please provide your complete output for this sub-task. Include:
1. Clear explanation of your approach
2. Code/specifications/tests as appropriate
3. Any dependencies or prerequisites
4. Next steps or follow-up work needed
EOF

    # Execute with opencode (using different model/agent preferences)
    # Note: Adjust model names based on your opencode configuration
    case "$agent" in
        gemini)
            opencode --files "$TASKS_FILE,design.md,requirements.md" \
                < "$RESULTS_DIR/${agent}_${subtask_id}_prompt.txt" \
                > "$output_file" 2> "$log_file" &
            ;;
        codex)
            opencode --files "$TASKS_FILE,design.md,requirements.md" \
                < "$RESULTS_DIR/${agent}_${subtask_id}_prompt.txt" \
                > "$output_file" 2> "$log_file" &
            ;;
        amp)
            opencode --files "$TASKS_FILE,design.md,requirements.md" \
                < "$RESULTS_DIR/${agent}_${subtask_id}_prompt.txt" \
                > "$output_file" 2> "$log_file" &
            ;;
        *)
            log_error "Unknown agent: $agent"
            return 1
            ;;
    esac
    
    local pid=$!
    echo "$pid" > "$RESULTS_DIR/${agent}_${subtask_id}.pid"
    
    log_info "[$agent] Sub-task $subtask_id running with PID $pid"
}

# Run agents in parallel
run_parallel_agents() {
    log_info "Starting parallel agent execution..."
    
    # Parse subtasks and assign to agents
    local num_agents=${#AGENTS[@]}
    local agent_idx=0
    
    # Read subtasks from JSON (using basic parsing, could use jq if available)
    if command -v jq &> /dev/null; then
        local num_subtasks=$(jq '.subtasks | length' "$RESULTS_DIR/subtasks.json")
        
        for i in $(seq 0 $((num_subtasks - 1))); do
            local subtask_id=$(jq -r ".subtasks[$i].id" "$RESULTS_DIR/subtasks.json")
            local subtask_desc=$(jq -r ".subtasks[$i].description" "$RESULTS_DIR/subtasks.json")
            local focus=$(jq -r ".subtasks[$i].focus" "$RESULTS_DIR/subtasks.json")
            
            local agent="${AGENTS[$agent_idx]}"
            execute_subtask "$agent" "$subtask_id" "$subtask_desc" "$focus"
            
            agent_idx=$(( (agent_idx + 1) % num_agents ))
        done
    else
        log_warning "jq not found, using simple parallel execution"
        # Fallback: just run 3 agents with generic tasks
        for i in 1 2 3; do
            local agent="${AGENTS[$((i-1))]}"
            execute_subtask "$agent" "$i" "Sub-task $i" "implementation"
        done
    fi
    
    log_info "Waiting for all agents to complete..."
    wait
    log_success "All parallel agents completed"
}

# Integrate results using opencode
integrate_results() {
    log_info "Integrating results from all agents..."
    
    # Collect all agent outputs
    local all_outputs=""
    for agent in "${AGENTS[@]}"; do
        for result_file in "$RESULTS_DIR/${agent}"_*.md; do
            if [ -f "$result_file" ]; then
                all_outputs+="
## Agent: $agent
$(cat "$result_file")

"
            fi
        done
    done
    
    # Create integration prompt
    cat > "$RESULTS_DIR/integration_prompt.txt" <<EOF
You are the integration coordinator. Multiple AI agents have worked on different sub-tasks in parallel.
Your job is to integrate their outputs into a cohesive, final deliverable.

Please:
1. Review all agent outputs below
2. Identify conflicts or overlaps
3. Merge the best ideas from each
4. Create a unified, production-ready output
5. Highlight any gaps that still need addressing

Agent Outputs:
$all_outputs
EOF

    # Run integration with opencode
    log_info "Running opencode integration..."
    opencode --files "$TASKS_FILE" < "$RESULTS_DIR/integration_prompt.txt" \
        > "$RESULTS_DIR/integrated_output.md" 2>&1
    
    log_success "Integration complete!"
    log_success "Final output: $RESULTS_DIR/integrated_output.md"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up temporary files..."
    # Keep results but clean up PIDs and prompts
    rm -f "$RESULTS_DIR"/*.pid
    rm -f "$RESULTS_DIR"/*_prompt.txt
}

# Main orchestration flow
main() {
    if [ $# -eq 0 ]; then
        echo "Usage: $0 \"Main task description\""
        echo ""
        echo "Example: $0 \"Build a web app with user authentication\""
        echo ""
        echo "This script will:"
        echo "  1. Split the task into sub-tasks"
        echo "  2. Run gemini, codex, and amp in parallel on sub-tasks"
        echo "  3. Integrate results using opencode"
        exit 1
    fi
    
    local main_task="$1"
    
    echo ""
    echo "========================================="
    echo "  Collaborative AI Orchestration"
    echo "========================================="
    echo ""
    
    trap cleanup EXIT
    
    check_prerequisites
    split_task "$main_task" "$TASKS_FILE"
    run_parallel_agents
    integrate_results
    
    echo ""
    log_success "Orchestration complete!"
    echo ""
    echo "Results available in: $RESULTS_DIR/"
    echo "  - subtasks.json: Task breakdown"
    echo "  - *_*.md: Individual agent outputs"
    echo "  - integrated_output.md: Final integrated result"
    echo ""
}

# Run main function with all arguments
main "$@"
