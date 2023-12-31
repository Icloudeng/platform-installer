# Check if Python or Python 3 is installed
if command -v python3 &>/dev/null; then
    echo "Python 3 is installed."
    python_version=$(python3 --version 2>&1)
    echo "Python 3 version: $python_version"
    python_command="python3"
elif command -v python &>/dev/null; then
    echo "Python is installed."
    python_version=$(python --version 2>&1)
    echo "Python version: $python_version"
    python_command="python"
else
    echo "Python or Python 3 is not installed. Please install Python and try again."
    exit 1
fi

# Check if the OS is Ubuntu
if [[ "$(uname)" == "Linux" ]]; then
    # Check if pip is installed
    sudo apt-get update
    sudo apt-get -y install python3-venv python3-pip jq
fi

# Create Python virtual environment
$python_command -m venv .venv

if [ ! -f ".venv/bin/activate" ]; then
    echo "Venv script not found"
    exit 1
fi

source .venv/bin/activate

# Upgrade pip
$python_command -m pip install --upgrade pip

# PIP if Ansible is already installed
echo "Pip, start installing..."
pip install ansible jmespath requests netaddr
pip install python-telegram-bot python-dotenv redis
pip SQLAlchemy sqlalchemy-utils psycopg2-binary pymysql

# Ansible dependecies
ansible-galaxy install -r scripts/requirements.yaml

export ANSIBLE_HOST_KEY_CHECKING="False"
export ANSIBLE_CONFIG="$(pwd)/ansible.cfg"

ansible_log_file="logs/ansible_log.txt"

# Create ansible log file if not exists
if [[ ! -f $ansible_log_file ]]; then
    touch "$ansible_log_file"
    echo "Created $ansible_log_file file."
fi

random_secret=$($python_command -c 'import secrets; print(secrets.token_hex(16))')

# ############### PYTHON FUNTIONS ###############
# Read variables from /root/.env variable and pass them to extra variable
getenv="$python_command lib/getenv.py"

# extract variable (format: %%variables%%)
extract_vars="$python_command lib/extract_vars.py"

# Decode Metadata and can pass key
get_decoded_metadata="$python_command lib/metadata.py"

# Publish ansible playbook logs to a partical channel
logs_publisher="$python_command lib/logs_publisher.py"

# Publish message to a partical channel
redis_publisher="$python_command lib/redis_publisher.py"

# Log in file
logger="$python_command lib/logger.py"

# External DB Params
external_db="$python_command lib/external_db.py"

# END ############### PYTHON FUNTIONS ############### END

# Get admin system email
admin_email=$([ -z "$($getenv ADMIN_SYSTEM_EMAIL)" ] && echo "admin@homelab.com" || echo "$($getenv ADMIN_SYSTEM_EMAIL)")
