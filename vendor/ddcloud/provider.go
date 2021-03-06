package ddcloud

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider creates the Dimension Data Cloud resource provider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		// Provider settings schema
		Schema: map[string]*schema.Schema{
			"region": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "",
				Description:   "The region code that identifies the target end-point for the Dimension Data CloudControl API.",
				ConflictsWith: []string{"cloudcontrol_endpoint"},
			},
			"cloudcontrol_endpoint": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "",
				Description:   "The base URL of a custom target end-point for the Dimension Data CloudControl API.",
				ConflictsWith: []string{"region"},
			},
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The user name used to authenticate to the Dimension Data CloudControl API (if not specified, then the MCP_USER environment variable will be used).",
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Default:     "",
				Description: "The password used to authenticate to the Dimension Data CloudControl API (if not specified, then the MCP_PASSWORD environment variable will be used).",
			},
			"retry_count": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     3,
				Description: "The maximum number of times to retry operations that fail due to network connectivity errors.",
			},
			"retry_delay": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     5,
				Description: "The number of seconds to delay between operation retries.",
			},
			"allow_server_reboot": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Allow rebooting of ddcloud_server instances (e.g. for adding / removing NICs)?",
			},
			"auto_create_tag_keys": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "When applying a tag, automatically create the corresponding tag key if it is not defined?",
			},
		},

		// Provider resource definitions
		ResourcesMap: map[string]*schema.Resource{
			// A network domain.
			"ddcloud_networkdomain": resourceNetworkDomain(),

			// A VLAN.
			"ddcloud_vlan": resourceVLAN(),

			// A server (virtual machine).
			"ddcloud_server": resourceServer(),

			// An additional network interface card (NIC) in a server.
			"ddcloud_server_nic": resourceServerNIC(),

			// A server anti-affinity rule.
			"ddcloud_server_anti_affinity": resourceServerAntiAffinityRule(),

			// A Network Address Translation (NAT) rule.
			"ddcloud_nat": resourceNAT(),

			// A firewall rule.
			"ddcloud_firewall_rule": resourceFirewallRule(),

			// An IP address list.
			"ddcloud_address_list": resourceAddressList(),

			// A port list.
			"ddcloud_port_list": resourcePortList(),

			// A VIP node.
			"ddcloud_vip_node": resourceVIPNode(),

			// A VIP pool.
			"ddcloud_vip_pool": resourceVIPPool(),

			// A VIP pool member (links pool, node, and optionally port).
			"ddcloud_vip_pool_member": resourceVIPPoolMember(),

			// A virtual listener is the top-level entity for load-balancing functionality.
			"ddcloud_virtual_listener": resourceVirtualListener(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			// A network domain.
			"ddcloud_networkdomain": dataSourceNetworkDomain(),

			// A virtual network (VLAN).
			"ddcloud_vlan": dataSourceVLAN(),
		},

		// Provider configuration
		ConfigureFunc: configureProvider,
	}
}

// Configure the provider.
// Returns the provider's compute API client.
func configureProvider(providerSettings *schema.ResourceData) (interface{}, error) {
	// Log provider version (for diagnostic purposes).
	log.Print("ddcloud provider version is " + ProviderVersion)

	region := strings.ToLower(
		providerSettings.Get("region").(string),
	)
	customEndPoint := providerSettings.Get("cloudcontrol_endpoint").(string)
	if region == "" && customEndPoint == "" {
		return nil, fmt.Errorf("Neither the 'region' nor the 'cloudcontrol_endpoint' provider properties were specified (the 'ddcloud' provider requires exactly one of these properties to be configured).")
	}

	username := providerSettings.Get("username").(string)
	if isEmpty(username) {
		username = os.Getenv("MCP_USER")
		if isEmpty(username) {
			return nil, fmt.Errorf("The 'username' property was not specified for the 'ddcloud' provider, and the 'MCP_USER' environment variable is not present. Please supply either one of these to configure the user name used to authenticate to Dimension Data CloudControl.")
		}
	}

	password := providerSettings.Get("password").(string)
	if isEmpty(password) {
		password = os.Getenv("MCP_PASSWORD")
		if isEmpty(password) {
			return nil, fmt.Errorf("The 'password' property was not specified for the 'ddcloud' provider, and the 'MCP_PASSWORD' environment variable is not present. Please supply either one of these to configure the password used to authenticate to Dimension Data CloudControl.")
		}
	}

	var client *compute.Client
	if region != "" {
		client = compute.NewClient(region, username, password)
	} else {
		client = compute.NewClientWithBaseAddress(customEndPoint, username, password)
	}

	// Configure retry, if required.
	var (
		retryCount int
		retryDelay int
	)
	value, ok := providerSettings.GetOk("retry_count")
	if ok {
		retryCount = value.(int)
	}

	// Override retry configuration with environment variables, if required.
	retryValue, err := strconv.Atoi(os.Getenv("MCP_MAX_RETRY"))
	if err == nil {
		retryCount = retryValue

		retryValue, err = strconv.Atoi(os.Getenv("MCP_RETRY_DELAY"))
		if err == nil {
			retryDelay = retryValue
		}
	}

	client.ConfigureRetry(retryCount, time.Duration(retryDelay)*time.Second)

	settings := &ProviderSettings{
		AllowServerReboots: providerSettings.Get("allow_server_reboot").(bool),
		AutoCreateTagKeys:  providerSettings.Get("auto_create_tag_keys").(bool),
	}

	// Override server reboot behaviour with environment variables, if required.
	allowRebootValue, err := strconv.ParseBool(os.Getenv("MCP_ALLOW_SERVER_REBOOT"))
	if err != nil {
		settings.AllowServerReboots = allowRebootValue
	}

	// Override automatic tag creation behaviour with environment variables, if required.
	autoCreateTagKeys, err := strconv.ParseBool(os.Getenv("MCP_AUTO_CREATE_TAG_KEYS"))
	if err != nil {
		settings.AutoCreateTagKeys = autoCreateTagKeys
	}

	provider := newProvider(client, settings)

	return provider, nil
}

// ProviderSettings represents the configuration for the ddcloud provider.
type ProviderSettings struct {
	// Allow rebooting of ddcloud_server instances if required during an update?
	//
	// For example, servers must be rebooted to add or remove network adapters.
	AllowServerReboots bool

	// Automatically create a tag keys if the tag being applied is not already defined?
	AutoCreateTagKeys bool
}

type providerState struct {
	// The CloudControl API client.
	apiClient *compute.Client

	// The provider settings.
	settings *ProviderSettings

	// Global lock for provider state.
	stateLock *sync.Mutex

	// Lock per network domain (prevent parallel provisioning for some resource types).
	domainLocks map[string]*sync.Mutex

	// Lock per server (prevent parallel provisioning operations for a given ddcloud_server resource).
	serverLocks map[string]*sync.Mutex
}

func newProvider(client *compute.Client, settings *ProviderSettings) *providerState {
	return &providerState{
		apiClient:   client,
		settings:    settings,
		stateLock:   &sync.Mutex{},
		domainLocks: make(map[string]*sync.Mutex),
		serverLocks: make(map[string]*sync.Mutex),
	}
}

// Client retrieves the CloudControl API client from provider state.
func (state *providerState) Client() *compute.Client {
	return state.apiClient
}

// Settings retrieves a copy of the provider settings.
func (state *providerState) Settings() ProviderSettings {
	return *state.settings // We return a copy because these settings should be read-only once the provider has been created.
}

// GetDomainLock retrieves the global lock for the specified network domain.
func (state *providerState) GetDomainLock(id string, ownerNameOrFormat string, formatArgs ...interface{}) *providerDomainLock {
	state.stateLock.Lock()
	defer state.stateLock.Unlock()

	lock, ok := state.domainLocks[id]
	if !ok {
		lock = &sync.Mutex{}
		state.domainLocks[id] = lock
	}

	return &providerDomainLock{
		domainID:  id,
		ownerName: fmt.Sprintf(ownerNameOrFormat, formatArgs...),
		lock:      lock,
	}
}

type providerDomainLock struct {
	domainID  string
	ownerName string
	lock      *sync.Mutex
}

// Acquire the network domain lock.
func (domainLock *providerDomainLock) Lock() {
	log.Printf("%s acquiring lock for domain '%s'...", domainLock.ownerName, domainLock.domainID)
	domainLock.lock.Lock()
	log.Printf("%s acquired lock for domain '%s'.", domainLock.ownerName, domainLock.domainID)
}

// Release the network domain lock.
func (domainLock *providerDomainLock) Unlock() {
	log.Printf("%s releasing lock for domain '%s'...", domainLock.ownerName, domainLock.domainID)
	domainLock.lock.Unlock()
	log.Printf("%s released lock for domain '%s'.", domainLock.ownerName, domainLock.domainID)
}

// GetServerLock retrieves the global lock for the specified server.
func (state *providerState) GetServerLock(id string, ownerNameOrFormat string, formatArgs ...interface{}) *providerServerLock {
	state.stateLock.Lock()
	defer state.stateLock.Unlock()

	lock, ok := state.serverLocks[id]
	if !ok {
		lock = &sync.Mutex{}
		state.serverLocks[id] = lock
	}

	return &providerServerLock{
		serverID:  id,
		ownerName: fmt.Sprintf(ownerNameOrFormat, formatArgs...),
		lock:      lock,
	}
}

type providerServerLock struct {
	serverID  string
	ownerName string
	lock      *sync.Mutex
}

// Acquire the server lock.
func (serverLock *providerServerLock) Lock() {
	log.Printf("%s acquiring lock for server '%s'...", serverLock.ownerName, serverLock.serverID)
	serverLock.lock.Lock()
	log.Printf("%s acquired lock for server '%s'.", serverLock.ownerName, serverLock.serverID)
}

// Release the server lock.
func (serverLock *providerServerLock) Unlock() {
	log.Printf("%s releasing lock for server '%s'...", serverLock.ownerName, serverLock.serverID)
	serverLock.lock.Unlock()
	log.Printf("%s released lock for server '%s'.", serverLock.ownerName, serverLock.serverID)
}
