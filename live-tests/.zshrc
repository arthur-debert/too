# Set vi mode
bindkey -v

# Set custom prompt
PS1="[too-test] $ "

# Create a wrapper function for too that always uses the correct data path
function too() {
    command too --data-path="${TODO_DB_PATH}" "$@"
}