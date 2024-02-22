package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/fatih/color"
	routeros "github.com/go-routeros/routeros"
	"github.com/spf13/cobra"
)

var (
	router_client       *routeros.Client
	keep_alive_interval time.Duration = 30 * time.Second

	address  string
	username string
	password string
	port     int
)

func init_routeros_connection() {
	full_address := fmt.Sprintf("%s:%d", address, port)
	var err error
	fmt.Println("Attempting to connect to RouterOS at address:", full_address)
	router_client, err = routeros.Dial(full_address, username, password)
	if err != nil {
		color.Red("Failed to connect to RouterOS: %v", err)
		os.Exit(1)
	}
	go keep_alive_connection()
}

func keep_alive_connection() {
	for {
		time.Sleep(keep_alive_interval)
		_, err := router_client.Run("/system/identity/print")
		if err != nil {
			log.Printf("Keep alive failed, attempting to reconnect: %v", err)
			init_routeros_connection()
		}
	}
}

func get_env_with_default(key string, default_val string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return default_val
}

func get_port_from_env_or_default(key string, default_val int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return default_val
}

func main() {
	address = get_env_with_default("MIKROTIK_ADDRESS", "192.168.88.1")
	username = get_env_with_default("MIKROTIK_USERNAME", "admin")
	password = get_env_with_default("MIKROTIK_PASSWORD", "")
	port = get_port_from_env_or_default("MIKROTIK_PORT", 8728)

	var root_cmd = &cobra.Command{
		Use:   "mikrotik-wifi",
		Short: "MikroTik WiFi management tool",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			init_routeros_connection()
		},
	}

	root_cmd.PersistentFlags().StringVarP(&address, "address", "a", address, "Address of the RouterOS device")
	root_cmd.PersistentFlags().StringVarP(&username, "username", "u", username, "Username for RouterOS authentication")
	root_cmd.PersistentFlags().StringVarP(&password, "password", "p", password, "Password for RouterOS authentication")
	root_cmd.PersistentFlags().IntVarP(&port, "port", "P", port, "Port of the RouterOS device API")

	root_cmd.AddCommand(list_cmd)
	root_cmd.AddCommand(create_cmd)
	root_cmd.AddCommand(update_cmd)
	root_cmd.AddCommand(remove_cmd)

	// Debug prints
	fmt.Printf("Debug - Address: %s, Username: %s, Password: %s, Port: %d\n", address, username, password, port)

	if err := root_cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var list_cmd = &cobra.Command{
	Use:   "list",
	Short: "List all Wi-Fi networks",
	Run: func(cmd *cobra.Command, args []string) {
		list_networks()
	},
}

var create_cmd = &cobra.Command{
	Use:   "create [ssid] [password]",
	Short: "Create a new Wi-Fi network",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		create_network(args[0], args[1])
	},
}

var update_cmd = &cobra.Command{
	Use:   "update [ssid] [property] [new_value]",
	Short: "Update an existing Wi-Fi network's ssid or password",
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		update_network(args[0], args[1], args[2])
	},
}

var remove_cmd = &cobra.Command{
	Use:   "remove [ssid]",
	Short: "Remove an existing Wi-Fi network",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remove_network(args[0])
	},
}

func list_networks() {
	resp, err := router_client.Run("/interface/wireless/print")
	if err != nil {
		color.Red("Failed to list networks: %v", err)
		return
	}
	for _, re := range resp.Re {
		ssid, _ := re.Map["ssid"]
		fmt.Println(ssid)
	}
}

func create_network(ssid_name, password string) {
	_, err := router_client.Run("/interface/wireless/print", "?ssid="+ssid_name)
	if err != nil {
		color.Red("Error checking for existing network: %v", err)
		return
	}

	_, err = router_client.Run("/interface/wireless/security-profiles/add", "=name="+ssid_name, "=wpa2-pre-shared-key="+password)
	if err != nil {
		color.Red("Failed to create security profile: %v", err)
		return
	}

	_, err = router_client.Run("/interface/wireless/add", "=name="+ssid_name, "=ssid="+ssid_name, "=security-profile="+ssid_name, "=master-interface=wlan1")
	if err != nil {
		color.Red("Failed to create network: %v", err)
		return
	}

	color.Green("Network created successfully: %s", ssid_name)
}

func update_network(ssid_name, property, new_value string) {
	_, err := router_client.Run("/interface/wireless/print", "?ssid="+ssid_name)
	if err != nil {
		color.Red("Error checking for existing network: %v", err)
		return
	}

	if property == "ssid" {
		_, err = router_client.Run("/interface/wireless/set", "?ssid="+ssid_name, "=ssid="+new_value)
	} else if property == "password" {
		_, err = router_client.Run("/interface/wireless/security-profiles/set", "?name="+ssid_name, "=wpa2-pre-shared-key="+new_value)
	}

	if err != nil {
		color.Red("Failed to update network: %v", err)
		return
	}

	color.Green("Network updated successfully: %s", ssid_name)
}

func remove_network(ssid_name string) {
	_, err := router_client.Run("/interface/wireless/remove", "?ssid="+ssid_name)
	if err != nil {
		color.Red("Failed to remove network: %v", err)
		return
	}

	color.Green("Network removed successfully: %s", ssid_name)
}
