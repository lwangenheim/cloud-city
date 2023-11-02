import pandas as pd
import random
import subprocess
import sys
import datetime


#Required Ascii art

print("""
                 ___
             ,--'___`--,
           ,'   / _ \   `,
          /   _/ / \ \_   \\
         '-,-'\ /   \ /`-,-`
 __.----==-|   |     |   |-==----.__
 `---==|---|   |     |   |---|==---'
         .-`-,/|     |\,-'-.
         |    \ `---' /    |
         `   \ \     / /   '
          |   \| --- |/   |
          `    |     |    '
           |   |     |   |
            \  `.   ,'  /
             \  |   |  /
              \ |   | /
               \`. .'/
               -o|_|o-
                 (_)

              Firespray
""")



# Function to get external IP
def get_external_ip(local_port):
    curl_command = f"curl -s -x socks5h://localhost:{local_port} https://ipinfo.io/ip"
    try:
        external_ip = subprocess.check_output(curl_command, shell=True).decode('utf-8').strip()
        return external_ip
    except subprocess.CalledProcessError:
        return None

# Function to log command details
def log_command_details(csv_log, droplet_id, external_ip, time_ran, nmap_args, status):
    with open(csv_log, 'a') as file:
        file.write(f"{droplet_id},{external_ip},{time_ran},\"{nmap_args}\",{status}\n")

# Main function to process nmap commands
def process_commands():
    csv_file = 'droplets.csv'  # Replace with your CSV file path
    log_file = 'nmap_command_log.csv'  # Log file to record the nmap command details
    df = pd.read_csv(csv_file, header=None, names=['droplet_id', 'ip_address', 'time_created', 'local_port'])

    while True:
        # Ask for nmap options from the user
        user_nmap_args = input("Enter nmap command options (or 'exit' to quit): ").strip()
        if user_nmap_args.lower() == 'exit':
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

        # Run the nmap command
        nmap_command = f"nmap -Pn --proxy socks4://127.0.0.1:{local_port} {user_nmap_args}"
        try:
            subprocess.run(nmap_command, shell=True, check=True)
            # Log the command as successful
            time_ran = datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')
            log_command_details(log_file, droplet_id, external_ip, time_ran, user_nmap_args, "Success")
            print(f"nmap command logged: {nmap_command}")
        except subprocess.CalledProcessError as e:
            # Log the command as failed
            log_command_details(log_file, droplet_id, external_ip, time_ran, user_nmap_args, "Failed")
            print("nmap scan failed.")
            print(e)

# Run the process_commands function
if __name__ == '__main__':
    process_commands()

