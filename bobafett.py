import pandas as pd
import random
import subprocess
import datetime
import shlex

# Function to get external IP
def get_external_ip(local_port):
    curl_command = f"curl -s -x socks5h://localhost:{local_port} https://ipinfo.io/ip"
    try:
        external_ip = subprocess.check_output(curl_command, shell=True).decode('utf-8').strip()
        return external_ip
    except subprocess.CalledProcessError:
        return None

# Function to log command details
def log_command_details(csv_log, droplet_id, external_ip, time_ran, command, status):
    with open(csv_log, 'a') as file:
        file.write(f"{droplet_id},{external_ip},{time_ran},\"{command}\",{status}\n")

# Function to update proxychains.conf
def update_proxychains_conf(local_port):
    conf_template = """
    strict_chain
    quiet_mode
    [ProxyList]
    socks4  127.0.0.1 {port}
    """
    with open('/etc/proxychains.conf', 'w') as conf_file:
        conf_file.write(conf_template.format(port=local_port))

# Main function to process commands
def process_commands():
    csv_file = 'droplets.csv'     # Replace with your CSV file path
    log_file = 'command_log.csv'  # Log file to record the command details
    df = pd.read_csv(csv_file, header=None, names=['droplet_id', 'ip_address', 'time_created', 'local_port'])

    while True:
        # Ask for a command from the user
        user_command = input("Enter the command to run (or 'exit' to quit): ").strip()
        if user_command.lower() == 'exit':
            print("Exiting...")
            break

        # Select a random tunnel
        random_tunnel = df.sample(n=1).iloc[0]
        local_port = random_tunnel['local_port']
        droplet_id = random_tunnel['droplet_id']

        # Get the external IP address
        external_ip = get_external_ip(local_port)
        if external_ip is None:
            print(f"Failed to obtain external IP for local port {local_port}.")
            continue
        else:
            print(f"The external IP address using local port {local_port} is: {external_ip}")

        # Update proxychains configuration
        update_proxychains_conf(local_port)

        # Execute the command through proxychains
        command_to_run = f"proxychains {user_command}"
        time_ran = datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')
        try:
            subprocess.run(shlex.split(command_to_run), check=True)
            # Log the command as successful
            log_command_details(log_file, droplet_id, external_ip, time_ran, user_command, "Success")
            print(f"Command logged and executed: {command_to_run}")
        except subprocess.CalledProcessError as e:
            # Log the command as failed
            log_command_details(log_file, droplet_id, external_ip, time_ran, user_command, "Failed")
            print("Command execution failed.")
            print(e)

# Run the process_commands function
if __name__ == '__main__':
    process_commands()

