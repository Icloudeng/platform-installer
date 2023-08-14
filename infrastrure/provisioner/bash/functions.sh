# Function to get the last n characters from a string
get_last_n_chars() {
    local input_string="$1"
    local n_chars="$2"

    # Get the length of the string
    local string_length=${#input_string}

    # Check if the string length is less than or equal to n_chars
    if ((string_length <= n_chars)); then
        echo "$input_string"
    else
        # Calculate the starting position for the last n_chars
        local start_position=$((string_length - n_chars))

        # Extract the last n_chars using parameter expansion
        echo "${input_string:$start_position}"
    fi
}

execute_ansible_playbook() {
    if [ -n "$reference" ]; then
        channel_publisher="$reference"
    else
        channel_publisher="$vm_ip"
    fi

    if [ -f "./private-key.pem" ]; then
        chmod 600 ./private-key.pem
        # Run the command in background
        ansible-playbook -u "$ansible_user" -i "$vm_ip," --private-key "./private-key.pem" "$playbook_path" --extra-vars "$ansible_extra_vars" --extra-vars "@scripts/variables.yaml" &
        # capture the process ID of the Ansible playbook command
        playbook_pid=$!

        # =====================

        # Run the logs exports
        $logs_exporter --channel "$channel_publisher" &
        # capture the process ID of the Logs exporter
        logs_exporter_pid=$!
    else
        # Run the command in background
        ansible-playbook -u "$ansible_user" -i "'$vm_ip,'" "$playbook_path" --extra-vars "$ansible_extra_vars" --extra-vars "@scripts/variables.yaml" &
        # capture the process ID of the Ansible playbook command
        playbook_pid=$!

        # =====================

        # Run the logs exports
        $logs_exporter --channel "$channel_publisher" &
        # capture the process ID of the Logs exporter
        logs_exporter_pid=$!
    fi

    # Wait for both commands to finish and capture their exit codes
    wait $playbook_pid
    # Capture the exit code of the Ansible playbook command
    playbook_result=$?

    # Wait just two second to make sure all python exportation netword process has been publish
    sleep 5

    # Kill the logs exporter process
    kill -SIGINT $logs_exporter_pid >/dev/null 2>&1

    if [ $playbook_result -eq 0 ]; then
        echo "Playbook succeeded!"
        ran_status="succeeded"
    else
        echo "Playbook failed!"
        ran_status="failed"
    fi
}

get_platform_static_secret() {
    static_secret_name="$platform-$(date +%Y-%m)"
    static_secret=$($python_command -c "import hashlib; print(hashlib.sha256('$static_secret_name'.encode()).hexdigest()[:32])")
    echo $static_secret
}
