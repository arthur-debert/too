#!/bin/bash

# JSON parser utility for too e2e tests
# Provides functions to extract information from too's JSON output

set -euo pipefail

# Extract basic information from a todo JSON object
# Usage: get_todo_field <json> <field>
# Fields: uid, parentId, text, completion_status
get_todo_field() {
    local json="$1"
    local field="$2"
    
    case "$field" in
        "uid")
            echo "$json" | jq -r '.uid'
            ;;
        "parentId")
            echo "$json" | jq -r '.parentId // ""'
            ;;
        "text")
            echo "$json" | jq -r '.text'
            ;;
        "completion_status")
            echo "$json" | jq -r '.statuses.completion'
            ;;
        *)
            echo "Unknown field: $field" >&2
            return 1
            ;;
    esac
}

# Get count of todos by status from AllTodos array
# Usage: count_todos_by_status <json_output> <status>
# Status: pending, done
count_todos_by_status() {
    local json_output="$1"
    local status="$2"
    
    echo "$json_output" | jq -r "[.AllTodos[] | select(.statuses.completion == \"$status\")] | length"
}

# Get total count of todos
# Usage: get_total_count <json_output>
get_total_count() {
    local json_output="$1"
    echo "$json_output" | jq -r '.TotalCount'
}

# Get done count
# Usage: get_done_count <json_output>  
get_done_count() {
    local json_output="$1"
    echo "$json_output" | jq -r '.DoneCount'
}

# Get todos by parent ID
# Usage: get_todos_by_parent <json_output> <parent_id>
# Use empty string for root todos
get_todos_by_parent() {
    local json_output="$1"
    local parent_id="$2"
    
    if [[ -z "$parent_id" ]]; then
        echo "$json_output" | jq -r '.AllTodos[] | select(.parentId == "")'
    else
        echo "$json_output" | jq -r ".AllTodos[] | select(.parentId == \"$parent_id\")"
    fi
}

# Get todo by text
# Usage: get_todo_by_text <json_output> <text>
get_todo_by_text() {
    local json_output="$1"
    local text="$2"
    
    echo "$json_output" | jq -r --arg text "$text" '.AllTodos[] | select(.text == $text)'
}

# Get affected todo UIDs as a space-separated list
# Usage: get_affected_todo_uids <json_output>
get_affected_todo_uids() {
    local json_output="$1"
    echo "$json_output" | jq -r '.AffectedTodos[]?.uid' | tr '\n' ' ' | sed 's/[[:space:]]*$//'
}

# Check if a todo exists with given text
# Usage: todo_exists <json_output> <text>
# Returns: 0 if exists, 1 if not
todo_exists() {
    local json_output="$1"
    local text="$2"
    
    local count
    # Use jq's --arg to safely pass the text parameter
    count=$(echo "$json_output" | jq -r --arg text "$text" '[.AllTodos[] | select(.text == $text)] | length')
    [[ "$count" -gt 0 ]]
}

# Validate parent-child relationship
# Usage: validate_parent_child <json_output> <parent_text> <child_text>
# Returns: 0 if relationship exists, 1 if not
validate_parent_child() {
    local json_output="$1"
    local parent_text="$2"
    local child_text="$3"
    
    local parent_uid child_parent_id
    parent_uid=$(echo "$json_output" | jq -r --arg text "$parent_text" '.AllTodos[] | select(.text == $text) | .uid')
    child_parent_id=$(echo "$json_output" | jq -r --arg text "$child_text" '.AllTodos[] | select(.text == $text) | .parentId')
    
    [[ "$parent_uid" == "$child_parent_id" ]]
}