# Cloud City

Welcome to Cloud City, an advanced proxy management tool designed to streamline and automate the process of utilizing cloud resources for security testing. Cloud City is inspired by the classic cloud proxy tool by Tom, reimagined and revamped to meet modern cybersecurity challenges. This utility is tailored for professionals who need to conduct rigorous testing against firewalls and ensure that security operations centers (SOCs) can detect their activity.

## Features

- **Dynamic Cloud Proxy**: Quickly creates and manages a series of Digital Ocean droplets to be used as SSH tunnel proxies.
- **Automatic CSV Generation**: Generates a CSV file listing all IP addresses used, facilitating easy log checking for clients.
- **Graceful Teardown**: Upon receiving a CTRL+C command, Cloud City will automatically destroy the droplets, terminate the SSH tunnels, and clean up your `known_hosts` to prevent future SSH conflicts.
- **Companion Scripts**: Two utility scripts, `lobot` and `bobafett`, enhance the functionality of Cloud City by providing specific proxy services for different use cases.

### Lobot

Lobot integrates seamlessly with `nmap`, providing you with a fresh, random IP address from your Digital Ocean droplets for each scan initiated, ensuring varied scan sources and improved test reliability.

### Bobafett

Bobafett is your go-to for proxychains. It enables you to run any command with a new IP each time, automatically updating your `proxychains.conf` to reflect the current active proxies.

Both scripts maintain a detailed log of every command executed and the corresponding IP address it was routed through, culminating in a comprehensive CSV report for evidential and auditing purposes.

## Getting Started

To leverage Cloud City, you'll need to have a Digital Ocean account and your SSH key ID readily available. Follow these simple steps to obtain your SSH key ID from Digital Ocean:

```bash
curl -X GET -H "Content-Type: application/json" \
    -H "Authorization: Bearer [Your_Digital_Ocean_Personal_Access_Token]" \
    "https://api.digitalocean.com/v2/account/keys"
```

*Note: Replace `[Your_Digital_Ocean_Personal_Access_Token]` with your actual Digital Ocean personal access token.*

## Prerequisites

- A Digital Ocean account.
- Digital Ocean Personal Access Token.
- SSH key added to your Digital Ocean account.

## Installation

Clone the repository to your local machine:

```bash
git clone [Repository-URL]
```

Navigate to the Cloud City directory:

```bash
cd cloud-city
```

Run the setup script (if available) or follow the installation instructions provided within the repository.

## Usage

Instructions on how to deploy droplets, initiate tunnels, and utilize `lobot` and `bobafett` will be provided in detail within the repository's usage section.

## Contributing

If you would like to contribute to the development of Cloud City, please read `CONTRIBUTING.md` for the process for submitting pull requests to us.

## Authors

* **Lee** - *Initial work* - Inspired by the original cloud proxy tool by Tom.

## License

This project is licensed under the MIT License - see the `LICENSE.md` file for details.

## Acknowledgments

- Hat tip to Tom for the original cloud proxy concept.
- Gratitude to the Digital Ocean community for their excellent API and resources.

---

*This README is a starting point for Cloud City's documentation. Further details and expansions will be made to ensure full coverage and clear instructions for all users.*

