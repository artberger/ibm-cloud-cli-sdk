package core_config

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/IBM-Bluemix/bluemix-cli-sdk/bluemix/configuration"
	"github.com/IBM-Bluemix/bluemix-cli-sdk/bluemix/models"
	"github.com/fatih/structs"
)

type raw map[string]interface{}

func (r raw) Marshal() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

func (r raw) Unmarshal(bytes []byte) error {
	return json.Unmarshal(bytes, r)
}

type BXConfigData struct {
	ConsoleEndpoint         string
	Region                  string
	RegionID                string
	RegionType              string
	IAMEndpoint             string
	IAMID                   string
	IAMToken                string
	IAMRefreshToken         string
	Account                 models.Account
	ResourceGroup           models.ResourceGroup
	PluginRepos             []models.PluginRepo
	Locale                  string
	Trace                   string
	ColorEnabled            string
	HTTPTimeout             int
	CLIInfoEndpoint         string
	CheckCLIVersionDisabled bool
	UsageStatsDisabled      bool
	raw                     raw
}

func NewBXConfigData() *BXConfigData {
	data := new(BXConfigData)
	data.raw = make(map[string]interface{})
	return data
}

func (data *BXConfigData) Marshal() ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}

func (data *BXConfigData) Unmarshal(bytes []byte) error {
	err := json.Unmarshal(bytes, data)
	if err != nil {
		return err
	}

	var raw raw
	err = json.Unmarshal(bytes, &raw)
	if err != nil {
		return err
	}
	data.raw = raw

	return nil
}

type bxConfigRepository struct {
	data      *BXConfigData
	persistor configuration.Persistor
	initOnce  *sync.Once
	lock      sync.RWMutex
	onError   func(error)
}

func createBluemixConfigFromPath(configPath string, errHandler func(error)) *bxConfigRepository {
	return createBluemixConfigFromPersistor(configuration.NewDiskPersistor(configPath), errHandler)
}

func createBluemixConfigFromPersistor(persistor configuration.Persistor, errHandler func(error)) *bxConfigRepository {
	return &bxConfigRepository{
		data:      NewBXConfigData(),
		persistor: persistor,
		initOnce:  new(sync.Once),
		onError:   errHandler,
	}
}

func (c *bxConfigRepository) init() {
	c.initOnce.Do(func() {
		err := c.persistor.Load(c.data)
		if err != nil {
			c.onError(err)
		}
	})
}

func (c *bxConfigRepository) read(cb func()) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	c.init()

	cb()
}

func (c *bxConfigRepository) write(cb func()) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.init()

	cb()

	c.data.raw = structs.Map(c.data)

	err := c.persistor.Save(c.data)
	if err != nil {
		c.onError(err)
	}
}

func (c *bxConfigRepository) writeRaw(cb func()) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.init()

	cb()

	err := c.persistor.Save(c.data.raw)
	if err != nil {
		c.onError(err)
	}
}

func (c *bxConfigRepository) ConsoleEndpoint() (endpoint string) {
	c.read(func() {
		endpoint = c.data.ConsoleEndpoint
	})
	return
}

func (c *bxConfigRepository) Region() (region models.Region) {
	c.read(func() {
		region = models.Region{
			ID:   c.data.RegionID,
			Name: c.data.Region,
			Type: c.data.RegionType,
		}
	})
	return
}

func (c *bxConfigRepository) CloudName() string {
	regionID := c.Region().ID
	if regionID == "" {
		return ""
	}

	splits := strings.Split(regionID, ":")
	if len(splits) != 3 {
		return ""
	}

	customer := splits[0]
	if customer != "ibm" {
		return customer
	}

	deployment := splits[1]
	switch {
	case deployment == "yp":
		return "bluemix"
	case strings.HasPrefix(deployment, "ys"):
		return "staging"
	default:
		return ""
	}
}

func (c *bxConfigRepository) CloudType() string {
	return c.Region().Type
}

func (c *bxConfigRepository) IAMEndpoint() (endpoint string) {
	c.read(func() {
		endpoint = c.data.IAMEndpoint
	})
	return
}

func (c *bxConfigRepository) IAMID() string {
	return NewIAMTokenInfo(c.IAMToken()).IAMID
}

func (c *bxConfigRepository) IAMToken() (token string) {
	c.read(func() {
		token = c.data.IAMToken
	})
	return
}

func (c *bxConfigRepository) IAMRefreshToken() (token string) {
	c.read(func() {
		token = c.data.IAMRefreshToken
	})
	return
}

func (c *bxConfigRepository) Account() (account models.Account) {
	c.read(func() {
		account = c.data.Account
	})
	return
}

func (c *bxConfigRepository) HasAccount() bool {
	return c.Account().GUID != ""
}

func (c *bxConfigRepository) IMSAccountID() string {
	return NewIAMTokenInfo(c.IAMToken()).Accounts.IMSAccountID
}

func (c *bxConfigRepository) ResourceGroup() (group models.ResourceGroup) {
	c.read(func() {
		group = c.data.ResourceGroup
	})
	return
}

func (c *bxConfigRepository) HasResourceGroup() (hasGroup bool) {
	c.read(func() {
		hasGroup = c.data.ResourceGroup.GUID != "" && c.data.ResourceGroup.Name != ""
	})
	return
}

func (c *bxConfigRepository) PluginRepos() (repos []models.PluginRepo) {
	c.read(func() {
		repos = c.data.PluginRepos
	})
	return
}

func (c *bxConfigRepository) PluginRepo(name string) (models.PluginRepo, bool) {
	for _, r := range c.PluginRepos() {
		if strings.EqualFold(r.Name, name) {
			return r, true
		}
	}
	return models.PluginRepo{}, false
}

func (c *bxConfigRepository) Locale() (locale string) {
	c.read(func() {
		locale = c.data.Locale
	})
	return
}

func (c *bxConfigRepository) Trace() (trace string) {
	c.read(func() {
		trace = c.data.Trace
	})
	return
}

func (c *bxConfigRepository) ColorEnabled() (enabled string) {
	c.read(func() {
		enabled = c.data.ColorEnabled
	})
	return
}

func (c *bxConfigRepository) HTTPTimeout() (timeout int) {
	c.read(func() {
		timeout = c.data.HTTPTimeout
	})
	return
}

func (c *bxConfigRepository) CLIInfoEndpoint() (endpoint string) {
	c.read(func() {
		endpoint = c.data.CLIInfoEndpoint
	})

	return endpoint
}

func (c *bxConfigRepository) CheckCLIVersionDisabled() (disabled bool) {
	c.read(func() {
		disabled = c.data.CheckCLIVersionDisabled
	})
	return
}

func (c *bxConfigRepository) UsageStatsDisabled() (disabled bool) {
	c.read(func() {
		disabled = c.data.UsageStatsDisabled
	})
	return
}

func (c *bxConfigRepository) SetConsoleEndpoint(endpoint string) {
	c.write(func() {
		c.data.ConsoleEndpoint = endpoint
	})
}

func (c *bxConfigRepository) SetRegion(region models.Region) {
	c.write(func() {
		c.data.Region = region.Name
		c.data.RegionID = region.ID
		c.data.RegionType = region.Type
	})
}

func (c *bxConfigRepository) SetIAMEndpoint(endpoint string) {
	c.write(func() {
		c.data.IAMEndpoint = endpoint
	})
}

func (c *bxConfigRepository) SetIAMToken(token string) {
	c.writeRaw(func() {
		c.data.IAMToken = token
		c.data.raw["IAMToken"] = token
	})
}

func (c *bxConfigRepository) SetIAMRefreshToken(token string) {
	c.writeRaw(func() {
		c.data.IAMRefreshToken = token
		c.data.raw["IAMRefreshToken"] = token
	})
}

func (c *bxConfigRepository) SetAccount(account models.Account) {
	c.write(func() {
		c.data.Account = account
	})
}

func (c *bxConfigRepository) SetResourceGroup(group models.ResourceGroup) {
	c.write(func() {
		c.data.ResourceGroup = group
	})
}

func (c *bxConfigRepository) SetPluginRepo(pluginRepo models.PluginRepo) {
	c.write(func() {
		c.data.PluginRepos = append(c.data.PluginRepos, pluginRepo)
	})
}

func (c *bxConfigRepository) UnSetPluginRepo(repoName string) {
	c.write(func() {
		i := 0
		for ; i < len(c.data.PluginRepos); i++ {
			if strings.ToLower(c.data.PluginRepos[i].Name) == strings.ToLower(repoName) {
				break
			}
		}
		if i != len(c.data.PluginRepos) {
			c.data.PluginRepos = append(c.data.PluginRepos[:i], c.data.PluginRepos[i+1:]...)
		}
	})
}

func (c *bxConfigRepository) SetHTTPTimeout(timeout int) {
	c.write(func() {
		c.data.HTTPTimeout = timeout
	})
}

func (c *bxConfigRepository) SetCheckCLIVersionDisabled(disabled bool) {
	c.write(func() {
		c.data.CheckCLIVersionDisabled = disabled
	})
}

func (c *bxConfigRepository) SetCLIInfoEndpoint(endpoint string) {
	c.write(func() {
		c.data.CLIInfoEndpoint = endpoint
	})
}

func (c *bxConfigRepository) SetUsageStatsDisabled(disabled bool) {
	c.write(func() {
		c.data.UsageStatsDisabled = disabled
	})
}

func (c *bxConfigRepository) SetColorEnabled(enabled string) {
	c.write(func() {
		c.data.ColorEnabled = enabled
	})
}

func (c *bxConfigRepository) SetLocale(locale string) {
	c.write(func() {
		c.data.Locale = locale
	})
}

func (c *bxConfigRepository) SetTrace(trace string) {
	c.write(func() {
		c.data.Trace = trace
	})
}

func (c *bxConfigRepository) ClearSession() {
	c.write(func() {
		c.data.IAMToken = ""
		c.data.IAMRefreshToken = ""
		c.data.Account = models.Account{}
		c.data.ResourceGroup = models.ResourceGroup{}
	})
}

func (c *bxConfigRepository) ClearAPICache() {
	c.write(func() {
		c.data.Region = ""
		c.data.RegionID = ""
		c.data.ConsoleEndpoint = ""
		c.data.IAMEndpoint = ""
	})
}
