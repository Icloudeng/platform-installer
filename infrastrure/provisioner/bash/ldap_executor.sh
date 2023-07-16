# Check if the Ansible playbook script exists
playbook_path="./scripts/platforms/$platform/ldap-playbook.yaml"

if [ -f "$playbook_path" ]; then
    echo "Ansible LDAP playbook script found: $playbook_path"
else
    echo "Ansible LDAP playbook script not found for platform: $platform"
    exit 1
fi

# Test ldap object
ldap_data=$($get_decoded_metadata --key "ldap" --metadata "$metadata")
if jq -e . >/dev/null 2>&1 <<<"$ldap_data"; then
    echo "Parsed LDAP JSON successfully"
else
    echo "Failed to parse LDAP JSON"
    exit 1
fi

# Overwrite global varaible
################ Ansible extra-vars ################
ansible_extra_vars="platform_metadata=$metadata platform_name=$platform admin_email=$admin_email"

# Get the last total ansible logs file line number (Before ansible execution)
logs_lines=$(wc -l <$ansible_log_file | tr -d '[:space:]')

# Run Ansible playbook (Function) -> (ran_status: succeeded | failed)
execute_ansible_playbook

# Get the ansible logs content from last run and pipe it to base64 (After ansible execution)
ansible_logs=$(tail -n +$logs_lines $ansible_log_file)
ansible_logs_4096=$(get_last_n_chars "$ansible_logs" 4096 | base64)

# Read and extract variables exposed from ansible logs
exposed_variables=$($extract_vars "$ansible_logs")

# Execute python notifier script
installer_details+="EXECUTION TYPE: LDAP\n\n$exposed_variables\n"
$python_command lib/notifier.py --logs "$ansible_logs_4096" --status "$ran_status" --details "$installer_details" --metadata "$metadata"
