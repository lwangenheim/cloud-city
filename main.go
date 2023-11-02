package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"bufio"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

// Constants and global variables
const (
	patEnvVar      = "DIGITALOCEAN_ACCESS_TOKEN" // Personal Access Token environment variable.
	dropletCSVFile = "droplets.csv"              // CSV file for droplet details.
	defaultImage   = "ubuntu-20-04-x64"          // Default image slug for droplet creation
	defaultSize    = "s-1vcpu-1gb"               // Default size slug for droplet
	defaultRegion  = "nyc1"                      // Default region for droplet creation
)

var (
	count     = 5                                // Amount of droplets to deploy
	imageSlug = defaultImage                     // Image slug to use for the droplet
	sizeSlug  = defaultSize                      // Size slug to use for the droplet
	regions   = []string{"nyc1", "lon1", "sgp1"} // List of regions
	sshKeyID  = 12345853                         // Your SSH key ID 
)

// DropletInfo holds the details for a DigitalOcean droplet.
type DropletInfo struct {
	ID           int
	IP           string
	CreatedAt    time.Time
	SSHTunnelCmd *exec.Cmd
	LocalPort    int
}

// getTokenFromEnv retrieves the DigitalOcean API token from the environment variable.
func getTokenFromEnv() string {
	pat := os.Getenv(patEnvVar)
	if pat == "" {
		log.Fatalf("Environment variable %s not set.\n", patEnvVar)
	}
	return pat
}

// getRandomRegion returns a random region from the list of available regions.
func getRandomRegion() string {
	return regions[rand.Intn(len(regions))]
}

// getAvailablePort finds a random open port between 8000 and 9000.
func getAvailablePort() int {
	for {
		port := rand.Intn(1000) + 8000
		conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			conn.Close()
			return port
		}
	}
}

// getSSHKeyPath retrieves the SSH private key path from the current user's home directory.
func getSSHKeyPath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Unable to find current user: %s", err)
	}
	return usr.HomeDir + "/.ssh/id_rsa" //Change to your private key location
}

// Establish SSH connection
func establishSSHTunnel(ip string, port int) (*exec.Cmd, error) {
	user := "root" // Default user for DigitalOcean droplets
	cmdStr := fmt.Sprintf("ssh -N -o ExitOnForwardFailure=yes -o StrictHostKeyChecking=no -D %d %s@%s", port, user, ip)
	cmdArgs := strings.Split(cmdStr, " ")

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("error setting up stderr for SSH tunnel: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("error starting SSH tunnel: %w", err)
	}

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Printf("SSH tunnel stderr: %s", scanner.Text())
		}
	}()

	go func() {
		if err := cmd.Wait(); err != nil {
			log.Printf("SSH tunnel process ended with error: %v", err)
		}
	}()

	return cmd, nil
}

// createDroplet function that initializes a droplet and returns its information.
func createDroplet(client *godo.Client, writer *csv.Writer) *DropletInfo {
    region := getRandomRegion()
    createRequest := &godo.DropletCreateRequest{
        Name:   "droplet-" + strconv.Itoa(time.Now().Nanosecond()),
        Region: region,
        Size:   sizeSlug,
        Image: godo.DropletCreateImage{
            Slug: imageSlug,
        },
        SSHKeys: []godo.DropletCreateSSHKey{{ID: sshKeyID}},
    }

    ctx := context.TODO()
    newDroplet, _, err := client.Droplets.Create(ctx, createRequest)
    if err != nil {
        log.Printf("Error creating droplet: %s\n", err)
        return nil
    }

    ip, err := waitForDropletIP(client, newDroplet.ID)
    if err != nil {
        log.Printf("Error retrieving droplet IP: %s\n", err)
        return nil
    }

    localPort := getAvailablePort()
    var cmd *exec.Cmd
    retryDuration := 60 * time.Second // Wait time before retrying SSH tunnel

    // Waiting for 30 seconds before attempting to establish the SSH tunnel
    time.Sleep(30 * time.Second)

    for attempts := 0; attempts < 3; attempts++ { // Try up to 3 times
        cmd, err = establishSSHTunnel(ip, localPort)
        if err != nil {
            if attempts < 2 { // Before the last attempt, log the error and wait
                log.Printf("Error establishing SSH tunnel, will retry in %s: %s\n", retryDuration, err)
                time.Sleep(retryDuration)
            } else {
                log.Printf("Error establishing SSH tunnel after retries: %s\n", err)
                return nil
            }
        } else {
            // Convert the creation time from string to time.Time
            createdAt, err := time.Parse(time.RFC3339, newDroplet.Created)
            if err != nil {
                log.Printf("Error parsing droplet creation time: %s\n", err)
                return nil
            }

            // SSH tunnel established successfully
            dropletInfo := &DropletInfo{
                ID:           newDroplet.ID,
                IP:           ip,
                CreatedAt:    createdAt,
                SSHTunnelCmd: cmd,
                LocalPort:    localPort,
            }
            // Log to CSV only after successful tunnel creation
            logDropletInfo(writer, dropletInfo)
            fmt.Printf("Created droplet: %s with ID: %d\nSSH dynamic tunnel established on local port %d to %s:22\n", ip, newDroplet.ID, localPort, ip)
            return dropletInfo // Exit the loop and function with success
        }
    }

    log.Println("Failed to establish SSH tunnel after multiple attempts.")
    return nil
}

// logDropletInfo logs the details of the created droplet to a CSV file.
func logDropletInfo(writer *csv.Writer, info *DropletInfo) {
    record := []string{
        strconv.Itoa(info.ID),
        info.IP,
        info.CreatedAt.Format(time.RFC3339),
        strconv.Itoa(info.LocalPort),
    }

    err := writer.Write(record)
    if err != nil {
        log.Fatalf("Error writing to CSV: %s\n", err)
    }
    
    // Make sure to flush the writer to ensure all buffered operations are applied.
    writer.Flush()

    err = writer.Error()
    if err != nil {
        log.Fatalf("Error flushing CSV writer: %s\n", err)
    }
}


// waitForDropletIP waits for the droplet to have a public IP assigned.
func waitForDropletIP(client *godo.Client, dropletID int) (string, error) {
	ctx := context.TODO()
	for {
		droplet, _, err := client.Droplets.Get(ctx, dropletID)
		if err != nil {
			return "", fmt.Errorf("error retrieving droplet: %w", err)
		}

		if droplet.Networks != nil {
			for _, v := range droplet.Networks.V4 {
				if v.Type == "public" {
					return v.IPAddress, nil
				}
			}
		}

		log.Println("Waiting for droplet IP address to be assigned...")
		time.Sleep(10 * time.Second)
	}
}


// handleInterrupts listens for interrupt signals to clean up before exiting.
func handleInterrupts(client *godo.Client, droplets []*DropletInfo) {
    c := make(chan os.Signal)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        fmt.Println("\nReceived an interrupt, cleaning up...")

        // Call the cleanup function here to destroy droplets and close tunnels
        cleanup(client, droplets)

        // Exit after cleanup
        os.Exit(1)
    }()
}

// cleanup will handle destroying the droplets and closing SSH tunnels.
func cleanup(client *godo.Client, droplets []*DropletInfo) {
    for _, droplet := range droplets {
        if droplet.SSHTunnelCmd != nil && droplet.SSHTunnelCmd.Process != nil {
            fmt.Printf("Killing SSH tunnel for droplet ID %d...\n", droplet.ID)
            if err := droplet.SSHTunnelCmd.Process.Kill(); err != nil {
                fmt.Fprintf(os.Stderr, "Error killing SSH tunnel process: %s\n", err)
            }
        }

        // Destroy the droplet
        _, err := client.Droplets.Delete(context.Background(), droplet.ID)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error destroying droplet ID %d: %s\n", droplet.ID, err)
        } else {
            fmt.Printf("Destroyed droplet ID %d.\n", droplet.ID)
        }
    }
}



func main() {
    token := getTokenFromEnv()
    oauthClient := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{
        AccessToken: token,
    }))
    client := godo.NewClient(oauthClient)

    file, err := os.OpenFile(dropletCSVFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
    if err != nil {
        log.Fatalf("Error opening CSV file: %s", err)
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    var droplets []*DropletInfo
    var wg sync.WaitGroup
    for i := 0; i < count; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            dropletInfo := createDroplet(client, writer)
            if dropletInfo != nil {
                droplets = append(droplets, dropletInfo)
            }
        }()
        time.Sleep(1 * time.Second) // Stagger the creation to avoid rate limits
    }
    wg.Wait()


        // Set up the interrupt handler here, after all droplets have been created.
    handleInterrupts(client, droplets)

    // Output the details of the droplets and SSH tunnels
    fmt.Println("Finished creating droplets and establishing SSH tunnels.")
    for _, droplet := range droplets {
        fmt.Printf("Droplet ID %d with IP %s is accessible through local port %d\n", droplet.ID, droplet.IP, droplet.LocalPort)
    }

    // Wait indefinitely until an interrupt (Ctrl+C) is received.
    select {}
}

